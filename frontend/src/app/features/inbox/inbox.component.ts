import { Component } from '@angular/core';

@Component({
  selector: 'app-inbox',
  standalone: true,
  template: `
    <div class="inbox-page">
      <div class="empty-state">
        <p>No conversations yet. Messages from connected marketplaces will appear here.</p>
      </div>
    </div>
  `,
  styles: [`
    .empty-state {
      text-align: center; padding: 60px 20px;
      background: #111d32; border: 1px dashed #1e3a5f;
      border-radius: 10px; color: #64748b;
    }
  `]
})
export class InboxComponent {}
