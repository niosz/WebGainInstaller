<script lang="ts">
  import { showLog, logMessages } from './stores';

  let logContainer: HTMLElement;

  $: if ($logMessages.length && logContainer) {
    requestAnimationFrame(() => {
      logContainer.scrollTop = logContainer.scrollHeight;
    });
  }
</script>

<div class="bg-gh-surface rounded-lg border border-gh-border overflow-hidden">
  <button
    class="w-full px-4 py-2 flex items-center justify-between border-b border-gh-border bg-gh-overlay
           hover:bg-gh-bg transition-colors cursor-pointer"
    on:click={() => showLog.update(v => !v)}
  >
    <span class="text-xs font-semibold text-gh-text-sec uppercase tracking-wider">Log</span>
    <svg class="w-4 h-4 text-gh-text-muted transition-transform" class:rotate-180={$showLog}
         viewBox="0 0 16 16" fill="currentColor">
      <path d="M4.427 7.427l3.396 3.396a.25.25 0 00.354 0l3.396-3.396A.25.25 0 0011.396 7H4.604a.25.25 0 00-.177.427z"/>
    </svg>
  </button>

  {#if $showLog}
    <div bind:this={logContainer}
         class="h-32 overflow-y-auto p-3 font-mono text-xs leading-relaxed text-gh-text-sec bg-gh-bg">
      {#each $logMessages as msg}
        <div class="whitespace-pre-wrap">{msg}</div>
      {/each}
      {#if $logMessages.length === 0}
        <div class="text-gh-text-muted">In attesa di avvio...</div>
      {/if}
    </div>
  {/if}
</div>
