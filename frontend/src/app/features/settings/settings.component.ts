import { Component } from '@angular/core';

@Component({
  selector: 'app-settings',
  standalone: true,
  template: `
    <div class="settings-page">
      <h2>Marketplace Connections</h2>
      <div class="marketplace-card">
        <div class="marketplace-info">
          <span class="marketplace-logo">&#128722;</span>
          <div>
            <strong>Prom.ua</strong>
            <p class="marketplace-desc">Connect your Prom.ua seller account</p>
          </div>
        </div>
        <button class="btn-outline">Connect</button>
      </div>
    </div>
  `,
  styles: [`
    h2 { margin: 0 0 24px; font-size: 20px; font-weight: 600; }
    .marketplace-card {
      background: #111d32; border: 1px solid #1e3a5f;
      border-radius: 10px; padding: 20px;
      display: flex; justify-content: space-between; align-items: center;
    }
    .marketplace-info { display: flex; align-items: center; gap: 16px; }
    .marketplace-logo { font-size: 32px; }
    .marketplace-desc { margin: 4px 0 0; font-size: 13px; color: #64748b; }
    .btn-outline {
      padding: 8px 20px; border: 1px solid #3b82f6;
      border-radius: 8px; color: #3b82f6; background: none;
      font-size: 14px; cursor: pointer; transition: all 0.15s;
    }
    .btn-outline:hover { background: rgba(59, 130, 246, 0.1); }
  `]
})
export class SettingsComponent {}
