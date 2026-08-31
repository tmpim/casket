<template>
  <span
    :title="version ? `New in Casket ${version}` : 'New in Casket'"
    class="inline-block whitespace-nowrap rounded bg-green-600 text-white px-2 py-1 text-xs font-semibold select-none"
  >
    {{ version ? `Casket ${version}` : 'New in Casket' }}
  </span>
</template>

<script setup lang="ts">
import { defineProps } from "vue";
const props = defineProps<{ version?: string; }>();
</script>
