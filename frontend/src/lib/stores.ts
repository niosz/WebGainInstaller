import { writable } from 'svelte/store';

export interface ModuleStatus {
  folderName: string;
  name: string;
  description: string;
  weight: number;
  status: 'pending' | 'installing' | 'completed' | 'error';
  error?: string;
}

export interface ProgressInfo {
  percentage: number;
  currentModule: string;
  currentStep: string;
  stepIndex: number;
  totalSteps: number;
}

export interface ProjectInfo {
  name: string;
  version: string;
}

export const projectInfo = writable<ProjectInfo>({ name: 'WebGain Installer', version: '0.0.0' });
export const modules = writable<ModuleStatus[]>([]);
export const progress = writable<ProgressInfo>({
  percentage: 0,
  currentModule: '',
  currentStep: '',
  stepIndex: 0,
  totalSteps: 0,
});
export const installState = writable<'idle' | 'running' | 'complete' | 'error'>('idle');
export const errorMessage = writable<string>('');
export const showLog = writable<boolean>(false);
export const logMessages = writable<string[]>([]);

export function addLog(msg: string) {
  logMessages.update(logs => {
    const timestamp = new Date().toLocaleTimeString('it-IT');
    return [...logs, `[${timestamp}] ${msg}`];
  });
}
