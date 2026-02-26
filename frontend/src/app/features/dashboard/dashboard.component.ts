import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { ApiService } from '../../core/services/api.service';

interface DashboardData {
  products: { total: number; active: number; draft: number; archived: number; errors: number };
  inbox: { total_conversations: number; new_conversations: number; unread_messages: number };
  reviews: { total: number; average_rating: number; new_count: number };
  marketplaces: { id: string; marketplace: string; is_active: boolean; last_sync_at: string | null; product_count: number }[];
  recent_activity: { type: string; title: string; details: string; created_at: string }[];
}

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, RouterLink],
  template: `
    <div class="dashboard">
      <div class="page-header">
        <h1>Dashboard</h1>
      </div>

      <!-- Main Stats Grid -->
      <div class="stats-grid">
        <a routerLink="/products" class="stat-card">
          <div class="stat-icon blue">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#3b82f6" stroke-width="2">
              <path d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4M4 7l8 4M4 7v10l8 4m0-10v10"/>
            </svg>
          </div>
          <div class="stat-info">
            <span class="stat-value">{{ data()?.products?.total || 0 }}</span>
            <span class="stat-label">Products</span>
          </div>
          <div class="stat-detail" *ngIf="data()?.products?.active">
            {{ data()!.products.active }} active
          </div>
        </a>

        <a routerLink="/products" class="stat-card">
          <div class="stat-icon green">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#10b981" stroke-width="2">
              <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
          </div>
          <div class="stat-info">
            <span class="stat-value">{{ data()?.products?.active || 0 }}</span>
            <span class="stat-label">Published</span>
          </div>
          <div class="stat-detail error" *ngIf="data()?.products?.errors">
            {{ data()!.products.errors }} errors
          </div>
        </a>

        <a routerLink="/inbox" class="stat-card" [class.has-badge]="(data()?.inbox?.unread_messages || 0) > 0">
          <div class="stat-icon purple">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#8b5cf6" stroke-width="2">
              <path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/>
            </svg>
          </div>
          <div class="stat-info">
            <span class="stat-value">{{ data()?.inbox?.unread_messages || 0 }}</span>
            <span class="stat-label">Unread Messages</span>
          </div>
          <div class="stat-detail" *ngIf="data()?.inbox?.new_conversations">
            {{ data()!.inbox.new_conversations }} new chats
          </div>
        </a>

        <a routerLink="/reviews" class="stat-card">
          <div class="stat-icon yellow">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#f59e0b" stroke-width="2">
              <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
            </svg>
          </div>
          <div class="stat-info">
            <span class="stat-value">
              <span *ngIf="data()?.reviews?.average_rating; else noRating">
                {{ data()!.reviews.average_rating | number:'1.1-1' }}
              </span>
              <ng-template #noRating>&mdash;</ng-template>
            </span>
            <span class="stat-label">Avg Rating</span>
          </div>
          <div class="stat-detail" *ngIf="data()?.reviews?.new_count">
            {{ data()!.reviews.new_count }} new reviews
          </div>
        </a>
      </div>

      <!-- Two-column layout -->
      <div class="dashboard-columns">
        <!-- Marketplace Status -->
        <div class="panel">
          <div class="panel-header">
            <h3>Marketplace Connections</h3>
            <a routerLink="/settings" class="panel-link">Manage →</a>
          </div>
          <div class="panel-body">
            <div *ngIf="!data()?.marketplaces?.length" class="empty-mini">
              <p>No marketplaces connected</p>
              <a routerLink="/settings" class="btn-outline-sm">Connect Now</a>
            </div>
            <div class="mp-item" *ngFor="let mp of data()?.marketplaces">
              <div class="mp-left">
                <div class="mp-dot" [class.active]="mp.is_active"></div>
                <div class="mp-info">
                  <span class="mp-name">{{ mp.marketplace | titlecase }}</span>
                  <span class="mp-sub">{{ mp.product_count }} products</span>
                </div>
              </div>
              <div class="mp-right">
                <span class="mp-sync" *ngIf="mp.last_sync_at">
                  Synced {{ formatTime(mp.last_sync_at) }}
                </span>
                <span class="mp-sync" *ngIf="!mp.last_sync_at">Never synced</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Recent Activity -->
        <div class="panel">
          <div class="panel-header">
            <h3>Recent Activity</h3>
          </div>
          <div class="panel-body">
            <div *ngIf="!data()?.recent_activity?.length" class="empty-mini">
              <p>No recent activity</p>
            </div>
            <div class="activity-item" *ngFor="let item of data()?.recent_activity">
              <div class="activity-dot" [class]="getActivityClass(item.type)"></div>
              <div class="activity-content">
                <span class="activity-title">{{ item.title }}</span>
                <span class="activity-time">{{ formatTime(item.created_at) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Quick Actions -->
      <div class="quick-actions">
        <h3>Quick Actions</h3>
        <div class="actions-grid">
          <a routerLink="/products" [queryParams]="{ action: 'new' }" class="action-card">
            <span class="action-icon">+</span>
            <span>Add Product</span>
          </a>
          <a routerLink="/inbox" class="action-card">
            <span class="action-icon">✉</span>
            <span>Open Inbox</span>
          </a>
          <a routerLink="/reviews" class="action-card">
            <span class="action-icon">★</span>
            <span>View Reviews</span>
          </a>
          <a routerLink="/settings" class="action-card">
            <span class="action-icon">⚙</span>
            <span>Settings</span>
          </a>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .dashboard { padding: 0; }
    .page-header { margin-bottom: 24px; }
    .page-header h1 { font-size: 24px; font-weight: 600; color: #e2e8f0; margin: 0; }

    /* Stats Grid */
    .stats-grid {
      display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px;
      margin-bottom: 24px;
    }
    .stat-card {
      background: #111d32; border: 1px solid #1e3a5f; border-radius: 12px;
      padding: 20px; display: flex; align-items: center; gap: 14px;
      text-decoration: none; color: inherit; transition: all 0.2s;
      position: relative;
    }
    .stat-card:hover { border-color: #2d5a8f; transform: translateY(-1px); }
    .stat-card.has-badge::after {
      content: ''; position: absolute; top: 12px; right: 12px;
      width: 8px; height: 8px; border-radius: 50%; background: #3b82f6;
      animation: pulse 2s infinite;
    }
    @keyframes pulse {
      0%, 100% { opacity: 1; }
      50% { opacity: 0.5; }
    }

    .stat-icon {
      width: 44px; height: 44px; border-radius: 10px;
      display: flex; align-items: center; justify-content: center;
      flex-shrink: 0;
    }
    .stat-icon.blue { background: rgba(59,130,246,0.1); }
    .stat-icon.green { background: rgba(16,185,129,0.1); }
    .stat-icon.purple { background: rgba(139,92,246,0.1); }
    .stat-icon.yellow { background: rgba(245,158,11,0.1); }

    .stat-info { display: flex; flex-direction: column; }
    .stat-value { font-size: 26px; font-weight: 700; color: #e2e8f0; line-height: 1; }
    .stat-label { font-size: 12px; color: #64748b; margin-top: 2px; }
    .stat-detail { margin-left: auto; font-size: 11px; color: #10b981; font-weight: 500; }
    .stat-detail.error { color: #ef4444; }

    /* Panels */
    .dashboard-columns {
      display: grid; grid-template-columns: 1fr 1fr; gap: 16px;
      margin-bottom: 24px;
    }
    .panel {
      background: #111d32; border: 1px solid #1e3a5f; border-radius: 12px;
      overflow: hidden;
    }
    .panel-header {
      display: flex; justify-content: space-between; align-items: center;
      padding: 16px 20px; border-bottom: 1px solid #1e3a5f;
    }
    .panel-header h3 { font-size: 15px; font-weight: 600; color: #e2e8f0; margin: 0; }
    .panel-link { color: #3b82f6; font-size: 13px; text-decoration: none; }
    .panel-link:hover { color: #60a5fa; }
    .panel-body { padding: 12px 20px; }

    .empty-mini { text-align: center; padding: 24px 0; color: #64748b; font-size: 13px; }
    .btn-outline-sm {
      display: inline-block; margin-top: 8px;
      padding: 5px 14px; border: 1px solid #3b82f6; border-radius: 6px;
      color: #3b82f6; font-size: 12px; text-decoration: none;
    }
    .btn-outline-sm:hover { background: rgba(59,130,246,0.1); }

    /* Marketplace items */
    .mp-item {
      display: flex; justify-content: space-between; align-items: center;
      padding: 10px 0; border-bottom: 1px solid #0a1628;
    }
    .mp-item:last-child { border-bottom: none; }
    .mp-left { display: flex; align-items: center; gap: 10px; }
    .mp-dot {
      width: 8px; height: 8px; border-radius: 50%; background: #64748b;
    }
    .mp-dot.active { background: #10b981; }
    .mp-name { color: #e2e8f0; font-size: 14px; font-weight: 500; display: block; }
    .mp-sub { color: #64748b; font-size: 11px; }
    .mp-sync { color: #64748b; font-size: 12px; }

    /* Activity */
    .activity-item {
      display: flex; gap: 10px; padding: 8px 0;
      border-bottom: 1px solid #0a1628;
    }
    .activity-item:last-child { border-bottom: none; }
    .activity-dot {
      width: 8px; height: 8px; border-radius: 50%; margin-top: 5px;
      flex-shrink: 0; background: #64748b;
    }
    .activity-dot.product { background: #3b82f6; }
    .activity-dot.sync { background: #10b981; }
    .activity-dot.review { background: #f59e0b; }
    .activity-dot.message { background: #8b5cf6; }
    .activity-content { display: flex; flex-direction: column; min-width: 0; }
    .activity-title {
      color: #cbd5e1; font-size: 13px; line-height: 1.3;
      white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
    }
    .activity-time { color: #475569; font-size: 11px; }

    /* Quick Actions */
    .quick-actions h3 {
      font-size: 15px; font-weight: 600; color: #e2e8f0;
      margin: 0 0 12px 0;
    }
    .actions-grid {
      display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px;
    }
    .action-card {
      background: #111d32; border: 1px solid #1e3a5f; border-radius: 10px;
      padding: 16px; text-align: center; text-decoration: none;
      color: #e2e8f0; font-size: 13px; transition: all 0.2s;
      display: flex; flex-direction: column; align-items: center; gap: 8px;
    }
    .action-card:hover { border-color: #3b82f6; background: #162032; }
    .action-icon { font-size: 20px; }

    @media (max-width: 900px) {
      .stats-grid { grid-template-columns: repeat(2, 1fr); }
      .dashboard-columns { grid-template-columns: 1fr; }
      .actions-grid { grid-template-columns: repeat(2, 1fr); }
    }
  `]
})
export class DashboardComponent implements OnInit {
  data = signal<DashboardData | null>(null);

  constructor(private api: ApiService) {}

  ngOnInit() {
    this.api.get<DashboardData>('/dashboard').subscribe({
      next: (d) => this.data.set(d),
      error: () => {}
    });
  }

  getActivityClass(type: string): string {
    if (type.includes('product')) return 'product';
    if (type.includes('sync') || type.includes('import')) return 'sync';
    if (type.includes('review')) return 'review';
    if (type.includes('message') || type.includes('conversation')) return 'message';
    return '';
  }

  formatTime(dateStr: string): string {
    if (!dateStr) return '';
    const d = new Date(dateStr);
    const now = new Date();
    const diff = now.getTime() - d.getTime();
    const mins = Math.floor(diff / 60000);
    if (mins < 1) return 'just now';
    if (mins < 60) return `${mins}m ago`;
    const hours = Math.floor(mins / 60);
    if (hours < 24) return `${hours}h ago`;
    const days = Math.floor(hours / 24);
    if (days < 7) return `${days}d ago`;
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  }
}
