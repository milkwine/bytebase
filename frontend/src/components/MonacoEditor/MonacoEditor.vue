<template>
  <div ref="editorContainerRef" v-bind="$attrs"></div>
  <BBSpin
    v-if="!isEditorLoaded"
    class="h-full w-full flex items-center justify-center"
  />
</template>

<script lang="ts" setup>
import {
  onMounted,
  ref,
  toRef,
  nextTick,
  watch,
  shallowRef,
  PropType,
  onBeforeUnmount,
} from "vue";
import type { editor as Editor } from "monaco-editor";
import { Database, SQLDialect, Table } from "@/types";
import { MonacoHelper, useMonaco } from "./useMonaco";
import { useLineDecorations } from "./lineDecorations";
import type { useLanguageClient } from "@sql-lsp/client";

const props = defineProps({
  value: {
    type: String,
    required: true,
  },
  dialect: {
    type: String as PropType<SQLDialect>,
    default: "mysql",
  },
  readonly: {
    type: Boolean,
    default: false,
  },
  autoFocus: {
    type: Boolean,
    default: true,
  },
});

const emit = defineEmits<{
  (e: "change", content: string): void;
  (e: "change-selection", content: string): void;
  (e: "save", content: string): void;
  (e: "ready"): void;
}>();

const sqlCode = toRef(props, "value");
const dialect = toRef(props, "dialect");
const readOnly = toRef(props, "readonly");
const monacoInstanceRef = ref<MonacoHelper>();
const editorContainerRef = ref<HTMLDivElement>();
// use shallowRef to avoid deep conversion which will cause page crash.
const editorInstanceRef = shallowRef<Editor.IStandaloneCodeEditor>();
const languageClientRef = ref<ReturnType<typeof useLanguageClient>>();

const isEditorLoaded = ref(false);

const initEditorInstance = () => {
  const { monaco, formatContent, setPositionAtEndOfLine } =
    monacoInstanceRef.value!;

  const model = monaco.editor.createModel(sqlCode.value, "sql");
  const editorInstance = monaco.editor.create(editorContainerRef.value!, {
    model,
    tabSize: 2,
    insertSpaces: true,
    autoClosingQuotes: "always",
    detectIndentation: false,
    folding: false,
    automaticLayout: true,
    readOnly: readOnly.value,
    minimap: {
      enabled: false,
    },
    wordWrap: "on",
    fixedOverflowWidgets: true,
    fontSize: 15,
    lineHeight: 24,
    scrollBeyondLastLine: false,
    padding: {
      top: 8,
      bottom: 8,
    },
    renderLineHighlight: "none",
    codeLens: false,
    scrollbar: {
      alwaysConsumeMouseWheel: false,
    },
  });

  // add `Format SQL` action into context menu
  editorInstance.addAction({
    id: "format-sql",
    label: "Format SQL",
    keybindings: [
      monaco.KeyMod.Alt | monaco.KeyMod.Shift | monaco.KeyCode.KeyF,
    ],
    contextMenuGroupId: "operation",
    contextMenuOrder: 1,
    run: () => {
      if (readOnly.value) {
        return;
      }
      formatContent(editorInstance, dialect.value);
      nextTick(() => setPositionAtEndOfLine(editorInstance));
    },
  });

  // typed something, change the text
  editorInstance.onDidChangeModelContent(() => {
    const value = editorInstance.getValue();
    emit("change", value);
  });

  // when editor change selection, emit change-selection event with selected text
  editorInstance.onDidChangeCursorSelection((e: any) => {
    const selectedText = editorInstance
      .getModel()
      ?.getValueInRange(e.selection) as string;
    emit("change-selection", selectedText);
  });

  editorInstance.onDidChangeCursorPosition(async (e: any) => {
    const { defineLineDecorations, disposeLineDecorations } =
      await useLineDecorations(editorInstance, e.position);
    // clear the old decorations
    disposeLineDecorations();

    // define the new decorations
    nextTick(async () => {
      await defineLineDecorations();
    });
  });

  editorInstance.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
    const value = editorInstance.getValue();
    emit("save", value);
  });

  return editorInstance;
};

onMounted(async () => {
  // Load monaco-editor and sql-lsp/client asynchronously.
  const [monacoHelper, { useLanguageClient }] = await Promise.all([
    useMonaco(),
    import("@sql-lsp/client"),
  ]);

  if (!editorContainerRef.value) {
    // Give up creating monaco editor if the component has been unmounted
    // very quickly.
    console.debug(
      "<MonacoEditor> has been unmounted before useMonaco is ready"
    );
    return;
  }

  const { setPositionAtEndOfLine } = monacoHelper;
  monacoInstanceRef.value = monacoHelper;

  const editorInstance = initEditorInstance();
  editorInstanceRef.value = editorInstance;

  const languageClient = useLanguageClient();
  languageClientRef.value = languageClient;
  languageClient.start();

  // set the editor focus when the tab is selected
  if (!readOnly.value && props.autoFocus) {
    editorInstance.focus();
    nextTick(() => setPositionAtEndOfLine(editorInstance));
  }

  isEditorLoaded.value = true;

  nextTick(() => {
    emit("ready");

    watch(dialect, () => languageClient.changeDialect(dialect.value), {
      immediate: true,
      // Delay the flush timing to ensure it performs after the language client started.
      flush: "post",
    });
  });
});

onBeforeUnmount(() => {
  editorInstanceRef.value?.dispose();
  monacoInstanceRef.value?.dispose();
  languageClientRef.value?.stop();
});

watch(
  () => readOnly.value,
  (readOnly) => {
    editorInstanceRef.value?.updateOptions({
      readOnly: readOnly,
    });
  },
  {
    deep: true,
    immediate: true,
  }
);

const getEditorContent = () => {
  return editorInstanceRef.value?.getValue();
};

const setEditorContent = (content: string) => {
  if (readOnly.value) {
    // workaround: setContent doesn't work in readonly mode
    // we temporarily set it to false
    editorInstanceRef.value?.updateOptions({
      readOnly: false,
    });
  }

  monacoInstanceRef.value?.setContent(editorInstanceRef.value!, content);

  if (readOnly.value) {
    // then set it back
    editorInstanceRef.value?.updateOptions({
      readOnly: true,
    });
  }
};

watch(
  () => props.value,
  (value) => {
    if (value !== getEditorContent()) {
      setEditorContent(value);
    }
  }
);

const getEditorContentHeight = () => {
  return editorInstanceRef.value?.getContentHeight();
};

const setEditorContentHeight = (height: number) => {
  editorContainerRef.value!.style.height = `${
    height ?? getEditorContentHeight()
  }px`;
};

const formatEditorContent = () => {
  if (readOnly.value) {
    return;
  }
  monacoInstanceRef.value?.formatContent(
    editorInstanceRef.value!,
    dialect.value
  );
  nextTick(() => {
    monacoInstanceRef.value?.setPositionAtEndOfLine(editorInstanceRef.value!);
    editorInstanceRef.value?.focus();
  });
};

const setEditorAutoCompletionContext = (
  databases: Database[],
  tables: Table[]
) => {
  languageClientRef.value?.changeSchema({
    databases: databases.map((db) => ({
      name: db.name,
      tables: tables
        .filter((table) => table.database.id === db.id)
        .map((table) => ({
          database: db.name,
          name: table.name,
          columns: table.columnList.map((col) => ({
            name: col.name,
          })),
        })),
    })),
  });
};

defineExpose({
  editorInstance: editorInstanceRef,
  formatEditorContent,
  getEditorContent,
  setEditorContent,
  getEditorContentHeight,
  setEditorContentHeight,
  setEditorAutoCompletionContext,
});
</script>

<style>
.monaco-editor .monaco-mouse-cursor-text {
  box-shadow: none !important;
}
.monaco-editor .scroll-decoration {
  display: none !important;
}
</style>
