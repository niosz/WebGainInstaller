<script lang="ts">
  export let status: 'pending' | 'installing' | 'completed' | 'error' = 'pending';
</script>

{#if status === 'pending'}
  <div class="w-5 h-5 rounded-full border-2 border-gh-border flex items-center justify-center flex-shrink-0">
    <div class="w-1.5 h-1.5 rounded-full bg-gh-text-muted"></div>
  </div>
{:else if status === 'installing'}
  <div class="w-5 h-5 flex items-center justify-center flex-shrink-0">
    <svg class="animate-spin w-5 h-5 text-gh-yellow" viewBox="0 0 24 24" fill="none">
      <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3" stroke-dasharray="31.4 31.4" stroke-linecap="round" />
    </svg>
  </div>
{:else if status === 'completed'}
  <div class="w-5 h-5 rounded-full bg-gh-green flex items-center justify-center flex-shrink-0">
    <svg class="w-3 h-3 text-white" viewBox="0 0 12 12" fill="none">
      <path d="M2 6L5 9L10 3" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
    </svg>
  </div>
{:else if status === 'error'}
  <div class="w-5 h-5 rounded-full bg-gh-red flex items-center justify-center flex-shrink-0">
    <svg class="w-3 h-3 text-white" viewBox="0 0 12 12" fill="none">
      <path d="M3 3L9 9M9 3L3 9" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
    </svg>
  </div>
{/if}
