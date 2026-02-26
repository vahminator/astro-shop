import { Component, inject, OnInit, signal, ElementRef, ViewChild, AfterViewChecked } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { InboxService, Conversation, MessageTemplate } from '../../core/services/inbox.service';
import { MarketplaceService } from '../../core/services/marketplace.service';

@Component({
  selector: 'app-inbox',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="inbox-page">
      <div class="inbox-layout">
        <!-- Conversation List -->
        <div class="conv-list-panel">
          <div class="panel-header">
            <h2>Inbox</h2>
            <button class="btn-icon" (click)="syncInbox()" [disabled]="syncing()" title="Sync from marketplace">
              <svg [class.spinning]="syncing()" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
              </svg>
            </button>
          </div>

          <div class="filter-tabs">
            <button [class.active]="statusFilter() === ''" (click)="filterByStatus('')">All</button>
            <button [class.active]="statusFilter() === 'new'" (click)="filterByStatus('new')">New</button>
            <button [class.active]="statusFilter() === 'in_progress'" (click)="filterByStatus('in_progress')">Active</button>
            <button [class.active]="statusFilter() === 'closed'" (click)="filterByStatus('closed')">Closed</button>
          </div>

          <div class="conv-list">
            @if (loading()) {
              <div class="loading-state">Loading...</div>
            } @else if (conversations().length === 0) {
              <div class="empty-state">
                <p>No conversations yet</p>
                <p class="hint">Sync your inbox to fetch messages</p>
              </div>
            } @else {
              @for (conv of conversations(); track conv.id) {
                <div class="conv-item" [class.active]="selectedConvId() === conv.id"
                     [class.unread]="conv.unread_count > 0"
                     (click)="selectConversation(conv)">
                  <div class="conv-avatar">{{ (conv.customer_name || '?').charAt(0).toUpperCase() }}</div>
                  <div class="conv-info">
                    <div class="conv-top">
                      <span class="conv-name">{{ conv.customer_name || 'Customer' }}</span>
                      <span class="conv-time">{{ formatTime(conv.last_message_at || conv.created_at) }}</span>
                    </div>
                    <div class="conv-preview">{{ conv.last_message || conv.subject || 'No messages' }}</div>
                    <div class="conv-meta">
                      <span class="mp-badge">{{ conv.marketplace }}</span>
                      @if (conv.unread_count > 0) {
                        <span class="unread-badge">{{ conv.unread_count }}</span>
                      }
                    </div>
                  </div>
                </div>
              }
            }
          </div>
        </div>

        <!-- Message Thread -->
        <div class="message-panel">
          @if (!selectedConvId()) {
            <div class="no-selection">
              <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="var(--text-muted)" stroke-width="1.5">
                <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
              </svg>
              <p>Select a conversation</p>
            </div>
          } @else {
            <div class="msg-header">
              <div class="msg-header-info">
                <h3>{{ selectedConvData()?.customer_name || 'Customer' }}</h3>
                <span class="mp-badge">{{ selectedConvData()?.marketplace }}</span>
              </div>
              <select [ngModel]="selectedConvData()?.status" (ngModelChange)="updateStatus($event)">
                <option value="new">New</option>
                <option value="in_progress">In Progress</option>
                <option value="waiting">Waiting</option>
                <option value="closed">Closed</option>
              </select>
            </div>

            <div class="messages-container" #messagesContainer>
              @if (messagesLoading()) {
                <div class="loading-state">Loading messages...</div>
              } @else {
                @for (msg of currentMessages(); track msg.id) {
                  <div class="message" [class.seller]="msg.sender === 'seller'" [class.customer]="msg.sender === 'customer'">
                    <div class="msg-bubble">
                      <div class="msg-body">{{ msg.body }}</div>
                      <div class="msg-time">{{ formatTime(msg.created_at) }}</div>
                    </div>
                  </div>
                }
                @if (currentMessages().length === 0) {
                  <div class="empty-messages">No messages yet</div>
                }
              }
            </div>

            <div class="reply-box">
              @if (templates().length > 0) {
                <div class="template-bar">
                  @for (t of templates(); track t.id) {
                    <button class="template-chip" (click)="replyText = t.body">{{ t.title }}</button>
                  }
                </div>
              }
              <div class="reply-row">
                <textarea [(ngModel)]="replyText" placeholder="Type your reply..." rows="2"
                          (keydown.enter)="onEnter($event)"></textarea>
                <button class="btn-send" (click)="sendReply()" [disabled]="!replyText.trim() || sending()">
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/>
                  </svg>
                </button>
              </div>
            </div>
          }
        </div>
      </div>
    </div>
  `,
  styles: [`
    .inbox-page { height: calc(100vh - 64px); }
    .inbox-layout { display: flex; height: 100%; }

    .conv-list-panel {
      width: 360px; border-right: 1px solid var(--border);
      display: flex; flex-direction: column; background: var(--bg-surface);
    }
    .panel-header {
      display: flex; justify-content: space-between; align-items: center;
      padding: 16px 20px; border-bottom: 1px solid var(--border);
    }
    .panel-header h2 { font-size: 18px; font-weight: 600; }
    .btn-icon {
      background: transparent; border: 1px solid var(--border); color: var(--text-secondary);
      width: 32px; height: 32px; border-radius: 6px; cursor: pointer;
      display: flex; align-items: center; justify-content: center;
    }
    .btn-icon:hover { background: var(--bg-surface-light); }
    .btn-icon:disabled { opacity: 0.5; }
    .spinning { animation: spin 1s linear infinite; }
    @keyframes spin { to { transform: rotate(360deg); } }

    .filter-tabs {
      display: flex; padding: 8px 12px; gap: 4px; border-bottom: 1px solid var(--border);
    }
    .filter-tabs button {
      padding: 6px 12px; border-radius: 6px; border: none;
      background: transparent; color: var(--text-muted); font-size: 13px; cursor: pointer;
    }
    .filter-tabs button:hover { color: var(--text-primary); background: var(--bg-surface-light); }
    .filter-tabs button.active { background: var(--accent-blue); color: white; }

    .conv-list { flex: 1; overflow-y: auto; }
    .conv-item {
      display: flex; gap: 12px; padding: 14px 16px; cursor: pointer;
      border-bottom: 1px solid var(--border);
    }
    .conv-item:hover { background: var(--bg-surface-light); }
    .conv-item.active { background: rgba(59,130,246,0.1); border-left: 3px solid var(--accent-blue); }
    .conv-item.unread { background: rgba(59,130,246,0.05); }

    .conv-avatar {
      width: 40px; height: 40px; border-radius: 50%; flex-shrink: 0;
      background: var(--bg-surface-light); display: flex; align-items: center;
      justify-content: center; font-weight: 600; font-size: 16px; color: var(--accent-blue);
    }
    .conv-info { flex: 1; min-width: 0; }
    .conv-top { display: flex; justify-content: space-between; margin-bottom: 4px; }
    .conv-name { font-weight: 500; font-size: 14px; }
    .conv-time { font-size: 11px; color: var(--text-muted); }
    .conv-preview {
      font-size: 13px; color: var(--text-muted); overflow: hidden;
      text-overflow: ellipsis; white-space: nowrap; margin-bottom: 4px;
    }
    .conv-meta { display: flex; align-items: center; gap: 8px; }
    .mp-badge {
      padding: 1px 6px; border-radius: 3px; font-size: 10px;
      font-weight: 600; text-transform: uppercase;
      background: rgba(59,130,246,0.15); color: var(--accent-blue);
    }
    .unread-badge {
      background: var(--accent-blue); color: white; font-size: 11px;
      font-weight: 600; padding: 1px 6px; border-radius: 10px;
    }

    .message-panel { flex: 1; display: flex; flex-direction: column; background: var(--bg-primary); }
    .no-selection {
      flex: 1; display: flex; flex-direction: column; align-items: center;
      justify-content: center; gap: 12px; color: var(--text-muted);
    }

    .msg-header {
      display: flex; justify-content: space-between; align-items: center;
      padding: 14px 20px; border-bottom: 1px solid var(--border); background: var(--bg-surface);
    }
    .msg-header-info { display: flex; align-items: center; gap: 10px; }
    .msg-header-info h3 { font-size: 16px; font-weight: 600; }
    .msg-header select {
      padding: 6px 10px; background: var(--bg-primary); border: 1px solid var(--border);
      border-radius: 6px; color: var(--text-primary); font-size: 13px;
    }

    .messages-container {
      flex: 1; overflow-y: auto; padding: 20px;
      display: flex; flex-direction: column; gap: 12px;
    }
    .loading-state, .empty-state, .empty-messages {
      text-align: center; padding: 32px; color: var(--text-muted);
    }
    .hint { font-size: 12px; margin-top: 4px; }

    .message { display: flex; }
    .message.customer { justify-content: flex-start; }
    .message.seller { justify-content: flex-end; }
    .msg-bubble { max-width: 70%; padding: 10px 14px; border-radius: 12px; font-size: 14px; line-height: 1.5; }
    .customer .msg-bubble { background: var(--bg-surface); border: 1px solid var(--border); border-bottom-left-radius: 4px; }
    .seller .msg-bubble { background: var(--accent-blue); color: white; border-bottom-right-radius: 4px; }
    .msg-body { white-space: pre-wrap; word-break: break-word; }
    .msg-time { font-size: 11px; margin-top: 4px; opacity: 0.6; text-align: right; }

    .reply-box { border-top: 1px solid var(--border); padding: 12px 16px; background: var(--bg-surface); }
    .template-bar { display: flex; gap: 6px; margin-bottom: 8px; flex-wrap: wrap; }
    .template-chip {
      padding: 4px 10px; border-radius: 12px; border: 1px solid var(--border);
      background: transparent; color: var(--text-secondary); font-size: 12px; cursor: pointer;
    }
    .template-chip:hover { background: var(--bg-surface-light); color: var(--text-primary); }
    .reply-row { display: flex; gap: 10px; align-items: flex-end; }
    .reply-row textarea {
      flex: 1; padding: 10px 12px; background: var(--bg-primary);
      border: 1px solid var(--border); border-radius: 8px;
      color: var(--text-primary); font-size: 14px; font-family: inherit;
      resize: none; outline: none;
    }
    .reply-row textarea:focus { border-color: var(--accent-blue); }
    .reply-row textarea::placeholder { color: var(--text-muted); }
    .btn-send {
      width: 40px; height: 40px; border-radius: 50%; flex-shrink: 0;
      background: var(--accent-blue); border: none; color: white;
      display: flex; align-items: center; justify-content: center; cursor: pointer;
    }
    .btn-send:hover { opacity: 0.9; }
    .btn-send:disabled { opacity: 0.4; cursor: not-allowed; }
  `]
})
export class InboxComponent implements OnInit, AfterViewChecked {
  @ViewChild('messagesContainer') messagesContainer!: ElementRef;

  private inboxService = inject(InboxService);
  private marketplaceService = inject(MarketplaceService);

  conversations = signal<Conversation[]>([]);
  selectedConvId = signal<string | null>(null);
  selectedConvData = signal<Conversation | null>(null);
  currentMessages = signal<any[]>([]);
  templates = signal<MessageTemplate[]>([]);
  connections = signal<any[]>([]);

  loading = signal(true);
  messagesLoading = signal(false);
  syncing = signal(false);
  sending = signal(false);
  statusFilter = signal('');
  replyText = '';
  private shouldScroll = false;

  ngOnInit() {
    this.loadConversations();
    this.inboxService.listTemplates().subscribe({ next: (t) => this.templates.set(t) });
    this.marketplaceService.list().subscribe({ next: (c) => this.connections.set(c) });
  }

  ngAfterViewChecked() {
    if (this.shouldScroll && this.messagesContainer) {
      this.messagesContainer.nativeElement.scrollTop = this.messagesContainer.nativeElement.scrollHeight;
      this.shouldScroll = false;
    }
  }

  loadConversations() {
    this.loading.set(true);
    this.inboxService.listConversations(this.statusFilter() || undefined).subscribe({
      next: (c) => { this.conversations.set(c); this.loading.set(false); },
      error: () => { this.conversations.set([]); this.loading.set(false); }
    });
  }

  filterByStatus(s: string) { this.statusFilter.set(s); this.loadConversations(); }

  selectConversation(conv: Conversation) {
    this.selectedConvId.set(conv.id);
    this.selectedConvData.set(conv);
    this.messagesLoading.set(true);
    this.inboxService.getConversation(conv.id).subscribe({
      next: (d) => {
        this.currentMessages.set(d.messages || []);
        this.messagesLoading.set(false);
        this.shouldScroll = true;
        this.conversations.update(cs => cs.map(c => c.id === conv.id ? { ...c, unread_count: 0 } : c));
      },
      error: () => { this.currentMessages.set([]); this.messagesLoading.set(false); }
    });
  }

  sendReply() {
    const id = this.selectedConvId();
    if (!id || !this.replyText.trim()) return;
    this.sending.set(true);
    const body = this.replyText.trim();
    this.replyText = '';
    this.inboxService.sendMessage(id, body).subscribe({
      next: () => {
        this.sending.set(false);
        this.currentMessages.update(m => [...m, {
          id: 'temp-' + Date.now(), sender: 'seller', body, is_read: true,
          created_at: new Date().toISOString(), external_id: ''
        }]);
        this.shouldScroll = true;
      },
      error: () => { this.sending.set(false); this.replyText = body; }
    });
  }

  updateStatus(status: string) {
    const id = this.selectedConvId();
    if (!id) return;
    this.inboxService.updateStatus(id, status).subscribe({
      next: () => {
        this.selectedConvData.update(c => c ? { ...c, status } : c);
        this.conversations.update(cs => cs.map(c => c.id === id ? { ...c, status } : c));
      }
    });
  }

  syncInbox() {
    const conn = this.connections()[0];
    if (!conn) return;
    this.syncing.set(true);
    this.inboxService.syncInbox(conn.id).subscribe({
      next: () => { this.syncing.set(false); setTimeout(() => this.loadConversations(), 3000); },
      error: () => this.syncing.set(false)
    });
  }

  onEnter(e: Event) {
    const ke = e as KeyboardEvent;
    if (!ke.shiftKey) { ke.preventDefault(); this.sendReply(); }
  }

  formatTime(d: string): string {
    if (!d) return '';
    const date = new Date(d);
    const diff = Date.now() - date.getTime();
    const mins = Math.floor(diff / 60000);
    if (mins < 1) return 'now';
    if (mins < 60) return `${mins}m`;
    const hours = Math.floor(diff / 3600000);
    if (hours < 24) return `${hours}h`;
    const days = Math.floor(diff / 86400000);
    if (days < 7) return `${days}d`;
    return date.toLocaleDateString('uk-UA', { day: '2-digit', month: '2-digit' });
  }
}
