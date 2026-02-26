import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MarketplaceService, MarketplaceConnection } from '../../core/services/marketplace.service';

@Component({
  selector: 'app-settings',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="settings-page">
      <h2>Marketplace Connections</h2>

      <!-- Existing connections -->
      @for (conn of connections(); track conn.id) {
        <div class="connection-card">
          <div class="connection-info">
            <div class="connection-header">
              <span class="marketplace-badge" [class]="conn.marketplace">
                {{ getMarketplaceName(conn.marketplace) }}
              </span>
              <span class="status-badge" [class.active]="conn.is_active" [class.inactive]="!conn.is_active">
                {{ conn.is_active ? 'Active' : 'Inactive' }}
              </span>
            </div>
            @if (conn.shop_url) {
              <p class="connection-url">{{ conn.shop_url }}</p>
            }
            @if (conn.last_sync_at) {
              <p class="connection-sync">Last sync: {{ conn.last_sync_at | date:'medium' }}</p>
            }
          </div>
          <div class="connection-actions">
            <button class="btn-sm btn-outline" (click)="testConnection(conn)" [disabled]="testing()">
              {{ testing() && testingId() === conn.id ? 'Testing...' : 'Test' }}
            </button>
            <button class="btn-sm btn-danger" (click)="deleteConnection(conn)">Delete</button>
          </div>
        </div>
      }

      @if (testResult()) {
        <div class="test-result" [class.success]="testResult()!.success" [class.error]="!testResult()!.success">
          {{ testResult()!.message }}
        </div>
      }

      <!-- Add new connection -->
      @if (!showForm()) {
        <button class="btn-primary add-btn" (click)="showForm.set(true)">+ Connect Marketplace</button>
      } @else {
        <div class="connect-form">
          <h3>Connect Prom.ua</h3>
          @if (formError()) {
            <div class="error-message">{{ formError() }}</div>
          }
          <div class="form-group">
            <label for="apiKey">API Key</label>
            <input id="apiKey" type="password" [(ngModel)]="apiKey" name="apiKey"
                   placeholder="Your Prom.ua API key" class="input">
          </div>
          <div class="form-group">
            <label for="shopUrl">Shop URL (optional)</label>
            <input id="shopUrl" type="text" [(ngModel)]="shopUrl" name="shopUrl"
                   placeholder="https://my-shop.prom.ua" class="input">
          </div>
          <div class="form-actions">
            <button class="btn-primary" (click)="connectMarketplace()" [disabled]="connecting()">
              {{ connecting() ? 'Connecting...' : 'Connect' }}
            </button>
            <button class="btn-ghost" (click)="cancelForm()">Cancel</button>
          </div>
        </div>
      }
    </div>
  `,
  styles: [`
    h2 { margin: 0 0 24px; font-size: 20px; font-weight: 600; }
    h3 { margin: 0 0 16px; font-size: 16px; font-weight: 600; }

    .connection-card {
      background: #111d32;
      border: 1px solid #1e3a5f;
      border-radius: 10px;
      padding: 16px 20px;
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 12px;
    }

    .connection-header {
      display: flex;
      align-items: center;
      gap: 10px;
      margin-bottom: 6px;
    }

    .marketplace-badge {
      padding: 2px 10px;
      border-radius: 4px;
      font-size: 13px;
      font-weight: 600;
    }
    .marketplace-badge.prom {
      background: rgba(59, 130, 246, 0.15);
      color: #3b82f6;
    }

    .status-badge {
      font-size: 12px;
      padding: 2px 8px;
      border-radius: 4px;
    }
    .status-badge.active {
      background: rgba(16, 185, 129, 0.15);
      color: #10b981;
    }
    .status-badge.inactive {
      background: rgba(100, 116, 139, 0.15);
      color: #64748b;
    }

    .connection-url, .connection-sync {
      font-size: 13px;
      color: #64748b;
      margin: 2px 0 0;
    }

    .connection-actions {
      display: flex;
      gap: 8px;
    }

    .btn-sm {
      padding: 6px 14px;
      border-radius: 6px;
      font-size: 13px;
      cursor: pointer;
      border: none;
    }

    .btn-outline {
      background: none;
      border: 1px solid #1e3a5f;
      color: #94a3b8;
    }
    .btn-outline:hover { border-color: #3b82f6; color: #3b82f6; }
    .btn-outline:disabled { opacity: 0.5; cursor: not-allowed; }

    .btn-danger {
      background: rgba(239, 68, 68, 0.1);
      color: #ef4444;
    }
    .btn-danger:hover { background: rgba(239, 68, 68, 0.2); }

    .test-result {
      padding: 10px 14px;
      border-radius: 8px;
      font-size: 13px;
      margin-bottom: 16px;
    }
    .test-result.success {
      background: rgba(16, 185, 129, 0.1);
      border: 1px solid rgba(16, 185, 129, 0.3);
      color: #10b981;
    }
    .test-result.error {
      background: rgba(239, 68, 68, 0.1);
      border: 1px solid rgba(239, 68, 68, 0.3);
      color: #ef4444;
    }

    .add-btn { margin-top: 8px; }

    .btn-primary {
      padding: 8px 20px;
      background: #3b82f6;
      color: white;
      border: none;
      border-radius: 8px;
      font-size: 14px;
      font-weight: 500;
      cursor: pointer;
    }
    .btn-primary:hover { background: #2563eb; }
    .btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

    .btn-ghost {
      padding: 8px 20px;
      background: none;
      color: #94a3b8;
      border: none;
      border-radius: 8px;
      font-size: 14px;
      cursor: pointer;
    }
    .btn-ghost:hover { color: #e2e8f0; }

    .connect-form {
      background: #111d32;
      border: 1px solid #1e3a5f;
      border-radius: 10px;
      padding: 24px;
      margin-top: 16px;
    }

    .form-group {
      display: flex;
      flex-direction: column;
      gap: 6px;
      margin-bottom: 16px;
    }
    .form-group label { font-size: 13px; font-weight: 500; color: #94a3b8; }
    .input {
      padding: 10px 14px;
      background: #0a1628;
      border: 1px solid #1e3a5f;
      border-radius: 8px;
      color: #e2e8f0;
      font-size: 14px;
      outline: none;
    }
    .input:focus { border-color: #3b82f6; }
    .input::placeholder { color: #475569; }

    .form-actions { display: flex; gap: 10px; }

    .error-message {
      padding: 10px 14px;
      background: rgba(239, 68, 68, 0.1);
      border: 1px solid rgba(239, 68, 68, 0.3);
      border-radius: 8px;
      color: #ef4444;
      font-size: 13px;
      margin-bottom: 16px;
    }
  `]
})
export class SettingsComponent implements OnInit {
  connections = signal<MarketplaceConnection[]>([]);
  showForm = signal(false);
  connecting = signal(false);
  testing = signal(false);
  testingId = signal('');
  testResult = signal<{ success: boolean; message: string } | null>(null);
  formError = signal('');

  apiKey = '';
  shopUrl = '';

  constructor(private marketplaceService: MarketplaceService) {}

  ngOnInit(): void {
    this.loadConnections();
  }

  loadConnections(): void {
    this.marketplaceService.list().subscribe({
      next: (data) => this.connections.set(data),
      error: () => {},
    });
  }

  getMarketplaceName(type: string): string {
    const names: Record<string, string> = {
      prom: 'Prom.ua',
      rozetka: 'Rozetka',
      epicenter: 'Epicenter',
      olx: 'OLX',
    };
    return names[type] || type;
  }

  connectMarketplace(): void {
    if (!this.apiKey) {
      this.formError.set('API key is required');
      return;
    }
    this.connecting.set(true);
    this.formError.set('');

    this.marketplaceService.connect({
      marketplace: 'prom',
      api_key: this.apiKey,
      shop_url: this.shopUrl || undefined,
    }).subscribe({
      next: () => {
        this.loadConnections();
        this.cancelForm();
        this.connecting.set(false);
      },
      error: (err) => {
        this.formError.set(err.error?.error || 'Failed to connect');
        this.connecting.set(false);
      },
    });
  }

  testConnection(conn: MarketplaceConnection): void {
    this.testing.set(true);
    this.testingId.set(conn.id);
    this.testResult.set(null);

    this.marketplaceService.testConnection(conn.id).subscribe({
      next: (res) => {
        this.testResult.set(res);
        this.testing.set(false);
      },
      error: () => {
        this.testResult.set({ success: false, message: 'Test request failed' });
        this.testing.set(false);
      },
    });
  }

  deleteConnection(conn: MarketplaceConnection): void {
    this.marketplaceService.delete(conn.id).subscribe({
      next: () => this.loadConnections(),
    });
  }

  cancelForm(): void {
    this.showForm.set(false);
    this.apiKey = '';
    this.shopUrl = '';
    this.formError.set('');
  }
}
