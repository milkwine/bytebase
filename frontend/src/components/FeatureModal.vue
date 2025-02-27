<template>
  <BBModal
    :title="$t('subscription.disabled-feature')"
    @close="$emit('cancel')"
  >
    <div class="min-w-0 md:min-w-400">
      <div class="flex items-start space-x-2">
        <div class="flex items-center">
          <heroicons-solid:sparkles class="h-6 w-6 text-accent" />
        </div>
        <h3
          id="modal-headline"
          class="flex self-center text-lg leading-6 font-medium text-gray-900"
        >
          {{ $t(`subscription.features.${featureKey}.title`) }}
        </h3>
      </div>
      <div class="mt-5">
        <p class="whitespace-pre-wrap">
          {{ $t(`subscription.features.${featureKey}.desc`) }}
        </p>
      </div>
      <div class="mt-3">
        <p class="whitespace-pre-wrap">
          {{
            $t("subscription.required-plan-with-trial", {
              requiredPlan: $t(
                `subscription.plan.${planTypeToString(requiredPlan)}.title`
              ),
              startTrial: $t("subscription.trial-for-plan", {
                plan: $t(
                  `subscription.plan.${planTypeToString(
                    subscriptionStore.trialingPlan
                  )}.title`
                ),
              }).toLowerCase(),
            })
          }}*
        </p>
      </div>
      <div class="mt-7 flex justify-end space-x-2">
        <button
          type="button"
          class="btn-normal"
          @click.prevent="$emit('cancel')"
        >
          {{ $t("common.dismiss") }}
        </button>

        <button
          v-if="subscriptionStore.canTrial"
          type="button"
          class="btn-primary"
          @click.prevent="trialSubscription"
        >
          {{
            $t("subscription.start-n-days-trial", {
              days: subscriptionStore.trialingDays,
            })
          }}
        </button>
        <button v-else type="button" class="btn-primary" @click.prevent="ok">
          {{ $t("common.learn-more") }}
        </button>
      </div>
    </div>
  </BBModal>
</template>

<script lang="ts" setup>
import { PropType } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { useSubscriptionStore, pushNotification } from "@/store";
import { FeatureType, planTypeToString } from "@/types";

const props = defineProps({
  feature: {
    required: true,
    type: String as PropType<FeatureType>,
  },
});

const emit = defineEmits(["cancel"]);
const { t } = useI18n();
const router = useRouter();

const ok = () => {
  router.push({ name: "setting.workspace.subscription" });
};

const subscriptionStore = useSubscriptionStore();

const requiredPlan = subscriptionStore.getMinimumRequiredPlan(props.feature);

const featureKey = props.feature.split(".").join("-");

const trialSubscription = () => {
  subscriptionStore.trialSubscription().then(() => {
    pushNotification({
      module: "bytebase",
      style: "SUCCESS",
      title: t("common.success"),
      description: t("subscription.successfully-start-trial", {
        days: subscriptionStore.trialingDays,
      }),
    });
    emit("cancel");
  });
};
</script>

<style scoped>
@media (min-width: 768px) {
  .md\:min-w-400 {
    min-width: 400px;
  }
}
</style>
