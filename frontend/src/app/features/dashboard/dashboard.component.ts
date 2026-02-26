import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [RouterLink],
  template: `
    <div class="dashboard">
      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-icon blue">&#128230;</div>
          <div class="stat-info">
            <span class="stat-value">0</span>
            <span class="stat-label">Products</span>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon green">&#9989;</div>
          <div class="stat-info">
            <span class="stat-value">0</span>
            <span class="stat-label">Published</span>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon blue">&#128172;</div>
          <div class="stat-info">
            <span class="stat-value">0</span>
            <span class="stat-label">Messages</span>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon yellow">&#11088;</div>
          <div class="stat-info">
            <span class="stat-value">&mdash;</span>
            <span class="stat-label">Avg Rating</span>
          </div>
        </div>
      </div>
      <div class="empty-state">
        <p>Connect a marketplace to get started</p>
        <a routerLink="/settings" class="btn-outline">Go to Settings</a>
      </div>
    </div>
  `,
  styles: [`
    .stats-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
      gap: 16px;
      margin-bottom: 32px;
    }
    .stat-card {
      background: #111d32;
      border: 1px solid #1e3a5f;
      border-radius: 10px;
      padding: 20px;
      display: flex;
      align-items: center;
      gap: 16px;
    }
    .stat-icon {
      width: 44px; height: 44px;
      border-radius: 10px;
      display: flex; align-items: center; justify-content: center;
      font-size: 20px;
    }
    .stat-icon.blue { background: rgba(59, 130, 246, 0.1); }
    .stat-icon.green { background: rgba(16, 185, 129, 0.1); }
    .stat-icon.yellow { background: rgba(245, 158, 11, 0.1); }
    .stat-value { font-size: 24px; font-weight: 700; display: block; }
    .stat-label { font-size: 13px; color: #64748b; }
    .empty-state {
      text-align: center; padding: 60px 20px;
      background: #111d32; border: 1px dashed #1e3a5f;
      border-radius: 10px; color: #64748b;
    }
    .btn-outline {
      display: inline-block; margin-top: 12px;
      padding: 8px 20px; border: 1px solid #3b82f6;
      border-radius: 8px; color: #3b82f6;
      text-decoration: none; font-size: 14px;
      transition: all 0.15s;
    }
    .btn-outline:hover { background: rgba(59, 130, 246, 0.1); }
  `]
})
export class DashboardComponent {}
