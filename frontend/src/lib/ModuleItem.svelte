<script lang="ts">
  import StatusIcon from './StatusIcon.svelte';
  import type { ModuleStatus } from './stores';

  export let module: ModuleStatus;
</script>

<div class="flex items-center gap-3 px-4 py-3 border-b border-gh-border-m last:border-b-0
            transition-colors duration-200"
     class:bg-gh-overlay={module.status === 'installing'}>
  <StatusIcon status={module.status} />

  <div class="flex-1 min-w-0">
    <div class="text-sm font-medium text-gh-text truncate">
      {module.name}
    </div>
    <div class="text-xs text-gh-text-sec truncate">
      {module.description}
    </div>
  </div>

  <div class="flex-shrink-0 text-right">
    {#if module.status === 'pending'}
      <span class="text-xs text-gh-text-muted">In attesa</span>
    {:else if module.status === 'installing'}
      <span class="text-xs text-gh-yellow">In corso...</span>
    {:else if module.status === 'completed'}
      <span class="text-xs text-gh-green">Completato</span>
    {:else if module.status === 'error'}
      <span class="text-xs text-gh-red" title={module.error || ''}>Errore</span>
    {/if}
  </div>
</div>
