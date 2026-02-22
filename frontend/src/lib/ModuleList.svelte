<script lang="ts">
  import ModuleItem from './ModuleItem.svelte';
  import type { ModuleStatus } from './stores';

  export let modules: ModuleStatus[] = [];
</script>

<div class="bg-gh-surface rounded-lg border border-gh-border overflow-hidden">
  <div class="px-4 py-2.5 border-b border-gh-border bg-gh-overlay">
    <span class="text-xs font-semibold text-gh-text-sec uppercase tracking-wider">Moduli</span>
    <span class="text-xs text-gh-text-muted ml-2">({modules.length})</span>
  </div>

  <div class="max-h-[320px] overflow-y-auto">
    {#each modules as mod (mod.folderName)}
      <ModuleItem module={mod} />
    {/each}
    {#if modules.length === 0}
      <div class="px-4 py-6 text-center text-gh-text-muted text-sm">
        Nessun modulo caricato
      </div>
    {/if}
  </div>
</div>
