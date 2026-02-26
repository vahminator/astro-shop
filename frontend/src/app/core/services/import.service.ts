import { Injectable, inject, signal, computed } from '@angular/core';
import { ApiService } from './api.service';

export interface ImportStatus {
  total: number;
  imported: number;
  skipped: number;
  failed: number;
  status: 'idle' | 'running' | 'completed' | 'failed';
  error?: string;
}

@Injectable({ providedIn: 'root' })
export class ImportService {
  private api = inject(ApiService);

  importStatus = signal<ImportStatus>({
    total: 0, imported: 0, skipped: 0, failed: 0, status: 'idle'
  });

  isRunning = computed(() => this.importStatus().status === 'running');
  progress = computed(() => {
    const s = this.importStatus();
    if (s.total === 0) return 0;
    return Math.round(((s.imported + s.skipped + s.failed) / s.total) * 100);
  });

  private pollInterval: any = null;

  startImport(marketplaceId: string) {
    return this.api.post<{ message: string }>('/import/start', { marketplace_id: marketplaceId });
  }

  startPolling() {
    this.stopPolling();
    this.pollStatus();
    this.pollInterval = setInterval(() => this.pollStatus(), 2000);
  }

  stopPolling() {
    if (this.pollInterval) {
      clearInterval(this.pollInterval);
      this.pollInterval = null;
    }
  }

  private pollStatus() {
    this.api.get<ImportStatus>('/import/status').subscribe({
      next: (status) => {
        this.importStatus.set(status);
        if (status.status !== 'running') {
          this.stopPolling();
        }
      },
      error: () => this.stopPolling()
    });
  }
}
