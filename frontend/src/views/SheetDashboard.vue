<template>
  <div class="w-full h-full grid grid-cols-[212px_1fr] gap-0">
    <div
      class="w-full h-auto flex flex-col justify-start items-start px-4 pl-6 py-2 border-r"
    >
      <router-link
        v-for="nav in navigationList"
        :key="nav.path"
        :to="nav.path"
        class="text-base w-full p-2 px-3 rounded-lg mt-1 select-none hover:bg-gray-100"
        active-class="active-link"
        exact-active-class="active-link"
        >{{ nav.label }}</router-link
      >
    </div>
    <div
      class="w-full h-full flex flex-col justify-start items-start overflow-y-auto px-4 py-4"
    >
      <div class="w-full px-4 mb-2 flex flex-row justify-between items-center">
        <div class="flex flex-row justify-start items-center">
          <div class="grow flex flex-row justify-start items-center">
            <span class="text-sm mr-2 whitespace-nowrap"
              >{{ $t("common.project") }}:
            </span>
            <n-select
              v-model:value="projectSelectorValue"
              :consistent-menu-width="false"
              :options="projectSelectOptions"
            />
          </div>
          <div class="ml-4">
            <n-input
              v-model:value="sheetSearchValue"
              type="text"
              :placeholder="t('common.search')"
            >
              <template #prefix>
                <heroicons-outline:search class="w-4 h-auto text-gray-300" />
              </template>
            </n-input>
          </div>
          <div class="ml-4">
            <n-button
              v-show="shouldShowClearSearchBtn"
              text
              @click="handleClearSearchBtnClick"
            >
              {{ $t("common.clear-search") }}
            </n-button>
          </div>
        </div>
        <div>
          <n-button
            v-if="isDev() && selectedProject?.workflowType === 'VCS'"
            @click="handleSyncSheetFromVCS"
          >
            <heroicons-outline:refresh class="w-4 h-auto mr-1" />
            {{ $t("sheet.actions.sync-from-vcs") }}
          </n-button>
        </div>
      </div>
      <div class="w-full grid grid-cols-1">
        <div
          class="sheet-list-container text-sm text-gray-400"
          :class="currentSubPath"
        >
          <span
            v-for="header in getSheetTableHeaderLabelList()"
            :key="header.key"
            >{{ header.label }}</span
          >
        </div>
        <div
          v-if="state.isLoading"
          class="w-full flex flex-col py-6 justify-start items-center"
        >
          <span class="text-sm leading-6 text-gray-500">{{
            $t("sql-editor.loading-data")
          }}</span>
        </div>
        <div
          v-for="sheet in shownSheetList"
          :key="sheet.id"
          class="sheet-list-container text-sm cursor-pointer hover:bg-gray-100"
          :class="currentSubPath"
          @click="handleSheetClick(sheet)"
        >
          <span
            v-for="value in getSheetTableContentValueList(sheet)"
            :key="value.key"
            class="truncate w-5/6"
            >{{ value.value }}</span
          >
          <div class="flex flex-row justify-end items-center" @click.stop>
            <n-dropdown
              trigger="click"
              :options="getSheetDropDownOptions(sheet)"
              @select="(key: string) => handleDropDownActionBtnClick(key, sheet)"
            >
              <heroicons-outline:dots-horizontal
                class="w-6 h-auto border border-gray-300 bg-white p-1 rounded outline-none shadow"
              />
            </n-dropdown>
          </div>
        </div>
        <div
          v-show="!state.isLoading && shownSheetList.length === 0"
          class="w-full flex flex-col py-6 justify-start items-center"
        >
          <heroicons-outline:inbox class="w-12 h-auto text-gray-500" />
          <span class="text-sm leading-6 text-gray-500">{{
            $t("common.no-data")
          }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { last } from "lodash-es";
import { useDialog } from "naive-ui";
import { t } from "@/plugins/i18n";
import dayjs from "@/plugins/dayjs";
import { useCurrentUser, useProjectStore, useSheetStore } from "@/store";
import { Sheet, SheetCreate, SheetOrganizerUpsert } from "@/types";
import {
  getDefaultSheetPayload,
  isDev,
  isSheetWritable,
  sheetSlug,
} from "@/utils";

interface LocalState {
  isLoading: boolean;
  sheetList: Sheet[];
}

const route = useRoute();
const router = useRouter();
const dialog = useDialog();
const state = reactive<LocalState>({
  isLoading: true,
  sheetList: [],
});
const currentUser = useCurrentUser();
const projectStore = useProjectStore();
const sheetStore = useSheetStore();

const projectSelectorValue = ref("");
const sheetSearchValue = ref("");

const navigationList = computed(() => {
  const list = [
    {
      path: "/sheets/my",
      label: t("sheet.my-sheets"),
    },
    {
      path: "/sheets/shared",
      label: t("sheet.shared-with-me"),
    },
    {
      path: "/sheets/starred",
      label: t("common.starred"),
    },
  ];

  return list;
});

const shownSheetList = computed(() => {
  return state.sheetList
    .filter((sheet) => {
      let t = true;

      if (
        projectSelectorValue.value !== "" &&
        projectSelectorValue.value !== sheet.project.name
      ) {
        t = false;
      }
      if (sheetSearchValue.value !== "") {
        if (
          !sheet.name.includes(sheetSearchValue.value) &&
          !sheet.statement.includes(sheetSearchValue.value)
        ) {
          t = false;
        }
      }

      return t;
    })
    .sort((a, b) => b.updatedTs - a.updatedTs);
});

const projectList = computed(() => {
  return projectStore.getProjectListByUser(currentUser.value.id);
});

const selectedProject = computed(() => {
  for (const project of projectList.value) {
    if (project.name === projectSelectorValue.value) {
      return project;
    }
  }

  return null;
});

const projectSelectOptions = computed(() => {
  return [
    {
      label: t("common.all"),
      value: "",
    },
  ].concat(
    projectList.value.map((project) => {
      return {
        label: project.name,
        value: project.name,
      };
    })
  );
});

const shouldShowClearSearchBtn = computed(() => {
  return projectSelectorValue.value !== "" || sheetSearchValue.value !== "";
});

const currentSubPath = computed(() => {
  const { path } = route;
  return last(path.split("/")) || "";
});

const fetchSheetData = async () => {
  if (currentSubPath.value === "my") {
    state.sheetList = await sheetStore.fetchMySheetList();
  } else if (currentSubPath.value === "starred") {
    state.sheetList = await sheetStore.fetchStarredSheetList();
  } else if (currentSubPath.value === "shared") {
    state.sheetList = await sheetStore.fetchSharedSheetList();
  }
};

onMounted(async () => {
  await projectStore.fetchProjectListByUser({
    userId: currentUser.value.id,
  });
  await fetchSheetData();
  state.isLoading = false;
});

watch(
  () => route.path,
  async () => {
    await fetchSheetData();
  }
);

const handleSheetClick = (sheet: Sheet) => {
  router.push({
    name: "sql-editor.share",
    params: {
      sheetSlug: sheetSlug(sheet),
    },
  });
};

const handleClearSearchBtnClick = () => {
  projectSelectorValue.value = "";
  sheetSearchValue.value = "";
};

const handleSyncSheetFromVCS = () => {
  if (
    selectedProject.value === null ||
    selectedProject.value.workflowType !== "VCS"
  ) {
    return;
  }

  const selectedProjectId = selectedProject.value.id;
  const dialogInstance = dialog.create({
    title: t("sheet.hint-tips.confirm-to-sync-sheet"),
    type: "info",
    async onPositiveClick() {
      dialogInstance.closable = false;
      dialogInstance.loading = true;
      await sheetStore.syncSheetFromVCS(selectedProjectId);
      await fetchSheetData();
      dialogInstance.destroy();
    },
    positiveText: t("common.confirm"),
    showIcon: true,
    maskClosable: false,
  });
};

const handleDropDownActionBtnClick = async (key: string, sheet: Sheet) => {
  if (key === "delete") {
    const dialogInstance = dialog.create({
      title: t("sheet.hint-tips.confirm-to-delete-this-sheet"),
      type: "info",
      async onPositiveClick() {
        await sheetStore.deleteSheetById(sheet.id);
        dialogInstance.destroy();
      },
      onNegativeClick() {
        dialogInstance.destroy();
      },
      negativeText: t("common.cancel"),
      positiveText: t("common.confirm"),
      showIcon: true,
    });
  } else if (key === "star" || key === "unstar") {
    const sheetOrganizerUpsert: SheetOrganizerUpsert = {
      sheeId: sheet.id,
    };

    if (key === "star") {
      sheetOrganizerUpsert.starred = true;
    } else if (key === "unstar") {
      sheetOrganizerUpsert.starred = false;
    }

    await sheetStore.upsertSheetOrganizer(sheetOrganizerUpsert);
    await fetchSheetData();
  } else if (key === "duplicate") {
    const dialogInstance = dialog.create({
      title: t("sheet.hint-tips.confirm-to-duplicate-sheet"),
      type: "info",
      async onPositiveClick() {
        const sheetCreate: SheetCreate = {
          projectId: sheet.projectId,
          name: sheet.name,
          statement: sheet.statement,
          visibility: "PRIVATE",
          payload: getDefaultSheetPayload(),
        };
        if (sheet.databaseId) {
          sheetCreate.databaseId = sheet.databaseId;
        }
        await sheetStore.createSheet(sheetCreate);
        dialogInstance.destroy();
      },
      onNegativeClick() {
        dialogInstance.destroy();
      },
      negativeText: t("common.cancel"),
      positiveText: t("common.confirm"),
      showIcon: true,
    });
  }
};

const getSheetDropDownOptions = (sheet: Sheet) => {
  const options = [];

  if (sheet.starred) {
    options.push({
      key: "unstar",
      label: t("common.unstar"),
    });
  } else {
    options.push({
      key: "star",
      label: t("common.star"),
    });
  }

  if (currentSubPath.value === "my" || currentSubPath.value === "starred") {
    options.push({
      key: "delete",
      label: t("common.delete"),
    });
  } else if (currentSubPath.value === "shared") {
    const canDeleteSheet = isSheetWritable(sheet, currentUser.value);
    if (canDeleteSheet) {
      options.push({
        key: "delete",
        label: t("common.delete"),
      });
    }

    options.push({
      key: "duplicate",
      label: t("common.duplicate"),
    });
  }

  return options;
};

const getSheetTableHeaderLabelList = () => {
  const labelList = [
    {
      key: "name",
      label: t("common.name").toUpperCase(),
    },
    {
      key: "project",
      label: t("common.project").toUpperCase(),
    },
    {
      key: "visibility",
      label: t("common.visibility").toUpperCase(),
    },
  ];

  if (currentSubPath.value === "shared" || currentSubPath.value === "starred") {
    labelList.push({
      key: "creator",
      label: t("common.creator").toUpperCase(),
    });
  }

  labelList.push({
    key: "updated",
    label: t("common.updated-at").toUpperCase(),
  });

  return labelList;
};

const getSheetTableContentValueList = (sheet: Sheet) => {
  const valueList = [
    {
      key: "name",
      value: sheet.name,
    },
    {
      key: "project",
      value: sheet.project.name,
    },
    {
      key: "visibility",
      value: sheet.visibility,
    },
  ];

  if (currentSubPath.value === "shared" || currentSubPath.value === "starred") {
    valueList.push({
      key: "creator",
      value: sheet.creator.name,
    });
  }

  valueList.push({
    key: "updated",
    value: dayjs.duration(sheet.updatedTs * 1000 - Date.now()).humanize(true),
  });

  return valueList;
};
</script>

<style scoped>
.active-link {
  @apply bg-gray-100 text-accent;
}
.sheet-list-container {
  @apply w-full grid py-3 px-4 border-b text-sm leading-6 select-none;
}
.sheet-list-container.my {
  grid-template-columns: 2fr repeat(3, 1fr) 32px;
}
.sheet-list-container.shared,
.sheet-list-container.starred {
  grid-template-columns: 2fr repeat(4, 1fr) 32px;
}
</style>
