<script lang="ts">
  import { onMount } from 'svelte';
  import { ConfirmCancel, RunSetupSteps, GetEulaText } from '../wailsjs/go/main/App.js';
  import { EventsOn } from '../wailsjs/runtime/runtime.js';

  type Screen = 'intro' | 'eula' | 'loader';
  let screen: Screen = 'intro';
  let stepMessage: string = '';
  let eulaText: string = '';
  let fatalError: boolean = false;

  async function loadEula() {
    eulaText = await GetEulaText();
  }

  onMount(() => {
    EventsOn('setup:step', (msg: string) => {
      stepMessage = msg;
    });

    EventsOn('setup:done', () => {
      stepMessage = '';
    });

    EventsOn('setup:fatal', () => {
      fatalError = true;
      stepMessage = '';
    });
  });

  async function handleCancel() {
    await ConfirmCancel();
  }

  function handleAccept() {
    screen = 'loader';
    RunSetupSteps();
  }
</script>

{#if screen === 'intro'}
  <main class="video-container">
    <video
      autoplay
      playsinline
      on:ended={() => { loadEula(); screen = 'eula'; }}
      src="/intro.mp4"
    >
      <track kind="captions" />
    </video>
  </main>

{:else if screen === 'eula'}
  <main class="eula-container">
    <header>
      <img src="/setup.png" alt="WebGain" class="header-icon" />
      <h1>WebGain Installer EULA</h1>
    </header>

    <div class="eula-box">
      <pre>{eulaText}</pre>
    </div>

    <footer>
      <button class="btn btn-cancel" on:click={handleCancel}>ANNULLA</button>
      <button class="btn btn-accept" on:click={handleAccept}>ACCETTA</button>
    </footer>
  </main>

{:else if screen === 'loader'}
  <main class="video-container">
    {#if !fatalError}
      <video
        autoplay
        muted
        loop
        playsinline
        src="/loader.mp4"
      >
        <track kind="captions" />
      </video>
    {/if}

    {#if stepMessage}
      <div class="step-overlay">
        <span class="step-text">{stepMessage}</span>
      </div>
    {/if}
  </main>
{/if}

<style>
  .video-container {
    width: 100vw;
    height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #0d1117;
    overflow: hidden;
    margin: 0;
    padding: 0;
    position: relative;
  }

  video {
    width: 100%;
    height: 100%;
    object-fit: contain;
    background: #0d1117;
  }

  .step-overlay {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    pointer-events: none;
  }

  .step-text {
    font-family: 'MesloLGS NF', 'Cascadia Code', 'Consolas', monospace;
    font-size: 20px;
    font-weight: 600;
    color: #ffffff;
    text-shadow: 0 2px 8px rgba(0, 0, 0, 0.8), 0 0 20px rgba(0, 0, 0, 0.6);
    text-align: center;
    padding: 12px 24px;
  }

  .eula-container {
    width: 100vw;
    height: 100vh;
    display: flex;
    flex-direction: column;
    background: #0d1117;
    padding: 32px 40px;
    overflow: hidden;
  }

  header {
    display: flex;
    align-items: center;
    gap: 20px;
    margin-bottom: 24px;
    flex-shrink: 0;
  }

  .header-icon {
    height: 60px;
    width: auto;
    object-fit: contain;
  }

  h1 {
    font-family: 'MesloLGS NF', 'Cascadia Code', 'Consolas', monospace;
    font-size: 26px;
    font-weight: 700;
    color: #ffffff;
    margin: 0;
    letter-spacing: 0.5px;
  }

  .eula-box {
    flex: 1;
    min-height: 0;
    background: #161b22;
    border: 1px solid #30363d;
    border-radius: 8px;
    padding: 24px;
    overflow-y: auto;
    margin-bottom: 24px;
  }

  .eula-box pre {
    font-family: 'MesloLGS NF', 'Cascadia Code', 'Consolas', monospace;
    font-size: 12px;
    line-height: 1.7;
    color: #8b949e;
    margin: 0;
    white-space: pre-wrap;
    word-wrap: break-word;
  }

  footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    flex-shrink: 0;
  }

  .btn {
    font-family: 'MesloLGS NF', 'Cascadia Code', 'Consolas', monospace;
    font-size: 13px;
    font-weight: 600;
    padding: 10px 28px;
    border-radius: 6px;
    border: 1px solid transparent;
    cursor: pointer;
    transition: all 0.15s ease;
    letter-spacing: 0.5px;
  }

  .btn-cancel {
    background: transparent;
    color: #8b949e;
    border-color: #30363d;
  }

  .btn-cancel:hover {
    background: #161b22;
    color: #e6edf3;
    border-color: #8b949e;
  }

  .btn-accept {
    background: #238636;
    color: #ffffff;
    border-color: #238636;
  }

  .btn-accept:hover {
    background: #2ea043;
    border-color: #2ea043;
  }

  .btn:active {
    transform: scale(0.97);
  }

  .eula-box::-webkit-scrollbar {
    width: 8px;
  }

  .eula-box::-webkit-scrollbar-track {
    background: #0d1117;
    border-radius: 4px;
  }

  .eula-box::-webkit-scrollbar-thumb {
    background: #30363d;
    border-radius: 4px;
  }

  .eula-box::-webkit-scrollbar-thumb:hover {
    background: #484f58;
  }
</style>
