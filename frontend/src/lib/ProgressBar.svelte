<script lang="ts">
  export let percentage: number = 0;
  export let currentModule: string = '';
  export let currentStep: string = '';

  $: clampedPct = Math.min(100, Math.max(0, percentage));
  $: displayPct = clampedPct.toFixed(1);
</script>

<div class="w-full">
  <div class="flex justify-between items-baseline mb-2">
    <span class="text-gh-text text-sm font-semibold">{displayPct}%</span>
    {#if currentModule}
      <span class="text-gh-text-sec text-xs truncate ml-4">
        {currentModule}
        {#if currentStep}
          <span class="text-gh-text-muted">— {currentStep}</span>
        {/if}
      </span>
    {/if}
  </div>

  <div class="w-full h-4 bg-gh-progress-bg rounded-full overflow-hidden border border-gh-border-m">
    <div
      class="h-full rounded-full transition-all duration-500 ease-out"
      style="width: {clampedPct}%; background: linear-gradient(90deg, #238636 0%, #2ea043 50%, #3fb950 100%);"
    ></div>
  </div>
</div>
