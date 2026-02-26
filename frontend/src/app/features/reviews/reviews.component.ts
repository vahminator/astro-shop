import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ReviewService, Review, ReviewStats } from '../../core/services/review.service';
import { MarketplaceService } from '../../core/services/marketplace.service';

@Component({
  selector: 'app-reviews',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="reviews-page">
      <div class="page-header">
        <h1>Review Hub</h1>
        <div class="header-actions">
          <button class="btn btn-secondary" (click)="syncReviews()" [disabled]="syncing()">
            <span [class.spin]="syncing()">↻</span> Sync Reviews
          </button>
        </div>
      </div>

      <!-- Stats Cards -->
      <div class="stats-row" *ngIf="stats()">
        <div class="stat-card">
          <div class="stat-value">{{ stats()!.total_reviews }}</div>
          <div class="stat-label">Total Reviews</div>
        </div>
        <div class="stat-card">
          <div class="stat-value rating-value">
            <span class="stars">{{ getStars(stats()!.average_rating) }}</span>
            {{ stats()!.average_rating | number:'1.1-1' }}
          </div>
          <div class="stat-label">Average Rating</div>
        </div>
        <div class="stat-card accent-blue">
          <div class="stat-value">{{ stats()!.new_count }}</div>
          <div class="stat-label">New</div>
        </div>
        <div class="stat-card accent-green">
          <div class="stat-value">{{ stats()!.responded_count }}</div>
          <div class="stat-label">Responded</div>
        </div>
      </div>

      <!-- Rating Breakdown -->
      <div class="rating-breakdown" *ngIf="stats() && stats()!.total_reviews > 0">
        <div class="breakdown-bar" *ngFor="let r of ratingValues">
          <span class="breakdown-label">{{ r }}★</span>
          <div class="bar-track">
            <div class="bar-fill" [style.width.%]="getRatingPercent(r)"></div>
          </div>
          <span class="breakdown-count">{{ stats()!.rating_breakdown[r] || 0 }}</span>
        </div>
      </div>

      <!-- Filters -->
      <div class="filters-bar">
        <div class="filter-tabs">
          <button *ngFor="let s of statusFilters"
                  [class.active]="filterStatus === s.value"
                  (click)="filterStatus = s.value; loadReviews()">
            {{ s.label }}
          </button>
        </div>
        <div class="filter-right">
          <select [(ngModel)]="filterRating" (change)="loadReviews()">
            <option value="">All Ratings</option>
            <option *ngFor="let r of ratingValues" [value]="r">{{ r }} Stars</option>
          </select>
          <input type="text" placeholder="Search reviews..." [(ngModel)]="searchQuery" (keyup.enter)="loadReviews()">
        </div>
      </div>

      <!-- Reviews List -->
      <div class="reviews-list" *ngIf="!loading(); else loadingTpl">
        <div *ngIf="reviews().length === 0" class="empty-state">
          <p>No reviews found. Sync from your marketplace or add manually.</p>
        </div>

        <div class="review-card" *ngFor="let review of reviews()" [class.expanded]="expandedId === review.id">
          <div class="review-header" (click)="toggleExpand(review.id)">
            <div class="review-left">
              <div class="review-rating">
                <span class="stars">{{ getStars(review.rating) }}</span>
                <span class="rating-num" *ngIf="review.rating > 0">{{ review.rating }}/5</span>
                <span class="no-rating" *ngIf="review.rating === 0">No rating</span>
              </div>
              <div class="review-meta">
                <span class="author">{{ review.author_name }}</span>
                <span class="separator">·</span>
                <span class="marketplace-badge">{{ review.marketplace }}</span>
                <span class="separator" *ngIf="review.product_title">·</span>
                <span class="product-name" *ngIf="review.product_title">{{ review.product_title }}</span>
              </div>
            </div>
            <div class="review-right">
              <span class="status-badge" [class]="'status-' + review.status">{{ review.status }}</span>
              <span class="review-date">{{ formatDate(review.created_at) }}</span>
              <span class="expand-icon">{{ expandedId === review.id ? '▾' : '▸' }}</span>
            </div>
          </div>

          <div class="review-body">
            {{ review.body }}
          </div>

          <div class="review-expanded" *ngIf="expandedId === review.id">
            <!-- Existing responses -->
            <div class="responses" *ngIf="review.responses.length > 0">
              <div class="response-item" *ngFor="let resp of review.responses">
                <div class="response-header">
                  <span class="response-label">Your Reply</span>
                  <span class="response-date">{{ formatDate(resp.created_at) }}</span>
                </div>
                <div class="response-body">{{ resp.body }}</div>
              </div>
            </div>

            <!-- Reply form -->
            <div class="reply-form">
              <textarea [(ngModel)]="replyText" placeholder="Write a reply..." rows="3"></textarea>
              <div class="reply-actions">
                <div class="status-actions">
                  <button class="btn-sm" *ngIf="review.status !== 'resolved'"
                          (click)="markResolved(review)">
                    ✓ Mark Resolved
                  </button>
                  <button class="btn-sm btn-danger" (click)="deleteReview(review)">Delete</button>
                </div>
                <button class="btn btn-primary btn-sm" (click)="sendReply(review)" [disabled]="!replyText.trim() || replying()">
                  {{ replying() ? 'Sending...' : 'Send Reply' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <ng-template #loadingTpl>
        <div class="loading">Loading reviews...</div>
      </ng-template>

      <!-- Pagination -->
      <div class="pagination" *ngIf="totalPages > 1">
        <button (click)="goToPage(currentPage - 1)" [disabled]="currentPage === 1">← Prev</button>
        <span class="page-info">Page {{ currentPage }} of {{ totalPages }}</span>
        <button (click)="goToPage(currentPage + 1)" [disabled]="currentPage === totalPages">Next →</button>
      </div>
    </div>
  `,
  styles: [`
    .reviews-page { padding: 0; }

    .page-header {
      display: flex; justify-content: space-between; align-items: center;
      margin-bottom: 24px;
    }
    .page-header h1 { font-size: 24px; font-weight: 600; color: #e2e8f0; margin: 0; }
    .header-actions { display: flex; gap: 10px; }

    .btn { padding: 8px 16px; border-radius: 8px; border: none; cursor: pointer; font-size: 13px; font-weight: 500; transition: all 0.2s; }
    .btn-primary { background: #3b82f6; color: #fff; }
    .btn-primary:hover { background: #2563eb; }
    .btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
    .btn-secondary { background: #1e3a5f; color: #e2e8f0; }
    .btn-secondary:hover { background: #254a75; }
    .btn-secondary:disabled { opacity: 0.5; }

    .spin { display: inline-block; animation: spin 1s linear infinite; }
    @keyframes spin { to { transform: rotate(360deg); } }

    /* Stats */
    .stats-row {
      display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px;
      margin-bottom: 20px;
    }
    .stat-card {
      background: #111d32; border: 1px solid #1e3a5f; border-radius: 10px;
      padding: 20px; text-align: center;
    }
    .stat-value { font-size: 28px; font-weight: 700; color: #e2e8f0; }
    .stat-label { font-size: 12px; color: #64748b; margin-top: 4px; text-transform: uppercase; letter-spacing: 0.5px; }
    .rating-value { display: flex; align-items: center; justify-content: center; gap: 8px; }
    .accent-blue .stat-value { color: #3b82f6; }
    .accent-green .stat-value { color: #10b981; }
    .stars { color: #f59e0b; }

    /* Rating Breakdown */
    .rating-breakdown {
      background: #111d32; border: 1px solid #1e3a5f; border-radius: 10px;
      padding: 16px 20px; margin-bottom: 20px;
      display: flex; flex-direction: column; gap: 8px;
    }
    .breakdown-bar { display: flex; align-items: center; gap: 10px; }
    .breakdown-label { color: #f59e0b; font-size: 13px; width: 30px; text-align: right; }
    .bar-track {
      flex: 1; height: 8px; background: #0a1628; border-radius: 4px; overflow: hidden;
    }
    .bar-fill { height: 100%; background: #f59e0b; border-radius: 4px; transition: width 0.3s; }
    .breakdown-count { color: #64748b; font-size: 12px; width: 30px; }

    /* Filters */
    .filters-bar {
      display: flex; justify-content: space-between; align-items: center;
      margin-bottom: 16px; gap: 16px; flex-wrap: wrap;
    }
    .filter-tabs { display: flex; gap: 4px; }
    .filter-tabs button {
      padding: 6px 14px; border-radius: 6px; border: 1px solid #1e3a5f;
      background: transparent; color: #64748b; cursor: pointer; font-size: 13px;
      transition: all 0.2s;
    }
    .filter-tabs button.active { background: #3b82f6; color: #fff; border-color: #3b82f6; }
    .filter-tabs button:hover:not(.active) { background: #111d32; color: #e2e8f0; }
    .filter-right { display: flex; gap: 10px; align-items: center; }
    .filter-right select, .filter-right input {
      padding: 6px 12px; border-radius: 6px; border: 1px solid #1e3a5f;
      background: #111d32; color: #e2e8f0; font-size: 13px;
    }
    .filter-right input { width: 200px; }

    /* Review Cards */
    .reviews-list { display: flex; flex-direction: column; gap: 8px; }
    .review-card {
      background: #111d32; border: 1px solid #1e3a5f; border-radius: 10px;
      overflow: hidden; transition: border-color 0.2s;
    }
    .review-card:hover { border-color: #2d5a8f; }
    .review-card.expanded { border-color: #3b82f6; }

    .review-header {
      display: flex; justify-content: space-between; align-items: center;
      padding: 14px 18px; cursor: pointer;
    }
    .review-left { display: flex; flex-direction: column; gap: 4px; }
    .review-rating { display: flex; align-items: center; gap: 8px; }
    .rating-num { color: #e2e8f0; font-size: 13px; font-weight: 500; }
    .no-rating { color: #64748b; font-size: 13px; font-style: italic; }
    .review-meta { display: flex; align-items: center; gap: 6px; font-size: 12px; }
    .author { color: #e2e8f0; font-weight: 500; }
    .separator { color: #334155; }
    .marketplace-badge {
      background: #1e3a5f; color: #3b82f6; padding: 1px 6px; border-radius: 4px;
      font-size: 10px; font-weight: 600; text-transform: uppercase;
    }
    .product-name { color: #64748b; max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

    .review-right { display: flex; align-items: center; gap: 12px; }
    .status-badge {
      padding: 3px 10px; border-radius: 12px; font-size: 11px; font-weight: 600;
      text-transform: capitalize;
    }
    .status-new { background: rgba(59,130,246,0.15); color: #3b82f6; }
    .status-responded { background: rgba(245,158,11,0.15); color: #f59e0b; }
    .status-resolved { background: rgba(16,185,129,0.15); color: #10b981; }
    .review-date { color: #64748b; font-size: 12px; }
    .expand-icon { color: #64748b; font-size: 12px; }

    .review-body {
      padding: 0 18px 14px; color: #cbd5e1; font-size: 14px; line-height: 1.5;
      white-space: pre-wrap;
    }

    .review-expanded { padding: 0 18px 18px; }

    .responses { margin-bottom: 14px; }
    .response-item {
      background: #0a1628; border: 1px solid #1e3a5f; border-radius: 8px;
      padding: 12px; margin-bottom: 8px;
    }
    .response-header { display: flex; justify-content: space-between; margin-bottom: 6px; }
    .response-label { color: #10b981; font-size: 12px; font-weight: 600; }
    .response-date { color: #64748b; font-size: 11px; }
    .response-body { color: #cbd5e1; font-size: 13px; line-height: 1.5; white-space: pre-wrap; }

    .reply-form textarea {
      width: 100%; padding: 10px 14px; border-radius: 8px;
      border: 1px solid #1e3a5f; background: #0a1628; color: #e2e8f0;
      font-size: 13px; resize: vertical; font-family: inherit;
      box-sizing: border-box;
    }
    .reply-form textarea:focus { outline: none; border-color: #3b82f6; }
    .reply-actions {
      display: flex; justify-content: space-between; align-items: center;
      margin-top: 10px;
    }
    .status-actions { display: flex; gap: 8px; }
    .btn-sm { padding: 5px 12px; font-size: 12px; border-radius: 6px; border: 1px solid #1e3a5f; background: transparent; color: #e2e8f0; cursor: pointer; }
    .btn-sm:hover { background: #1e3a5f; }
    .btn-danger { border-color: #ef4444; color: #ef4444; }
    .btn-danger:hover { background: rgba(239,68,68,0.15); }
    .btn.btn-sm { padding: 5px 14px; }

    /* Empty, loading, pagination */
    .empty-state {
      text-align: center; padding: 60px 20px;
      background: #111d32; border: 1px dashed #1e3a5f;
      border-radius: 10px; color: #64748b;
    }
    .loading { text-align: center; padding: 40px; color: #64748b; }
    .pagination {
      display: flex; justify-content: center; align-items: center; gap: 16px;
      margin-top: 20px; padding: 16px 0;
    }
    .pagination button {
      padding: 6px 14px; border-radius: 6px; border: 1px solid #1e3a5f;
      background: #111d32; color: #e2e8f0; cursor: pointer; font-size: 13px;
    }
    .pagination button:disabled { opacity: 0.4; cursor: not-allowed; }
    .page-info { color: #64748b; font-size: 13px; }
  `]
})
export class ReviewsComponent implements OnInit {
  reviews = signal<Review[]>([]);
  stats = signal<ReviewStats | null>(null);
  loading = signal(true);
  syncing = signal(false);
  replying = signal(false);

  filterStatus = '';
  filterRating = '';
  searchQuery = '';
  currentPage = 1;
  totalPages = 1;
  expandedId: string | null = null;
  replyText = '';

  ratingValues = [5, 4, 3, 2, 1];

  statusFilters = [
    { label: 'All', value: '' },
    { label: 'New', value: 'new' },
    { label: 'Responded', value: 'responded' },
    { label: 'Resolved', value: 'resolved' },
  ];

  private connections: any[] = [];

  constructor(
    private reviewService: ReviewService,
    private marketplaceService: MarketplaceService
  ) {}

  ngOnInit() {
    this.loadStats();
    this.loadReviews();
    this.marketplaceService.list().subscribe(c => this.connections = c);
  }

  loadStats() {
    this.reviewService.stats().subscribe(s => this.stats.set(s));
  }

  loadReviews() {
    this.loading.set(true);
    const params: any = { page: this.currentPage, limit: 20 };
    if (this.filterStatus) params.status = this.filterStatus;
    if (this.filterRating) params.rating = parseInt(this.filterRating);
    if (this.searchQuery) params.search = this.searchQuery;

    this.reviewService.list(params).subscribe({
      next: (res) => {
        this.reviews.set(res.reviews);
        this.totalPages = Math.ceil(res.total / res.limit) || 1;
        this.loading.set(false);
      },
      error: () => this.loading.set(false)
    });
  }

  syncReviews() {
    if (this.connections.length === 0) return;
    this.syncing.set(true);
    this.reviewService.sync(this.connections[0].id).subscribe({
      next: () => {
        setTimeout(() => {
          this.syncing.set(false);
          this.loadReviews();
          this.loadStats();
        }, 2000);
      },
      error: () => this.syncing.set(false)
    });
  }

  toggleExpand(id: string) {
    this.expandedId = this.expandedId === id ? null : id;
    this.replyText = '';
  }

  sendReply(review: Review) {
    if (!this.replyText.trim()) return;
    this.replying.set(true);
    this.reviewService.reply(review.id, this.replyText).subscribe({
      next: (updated) => {
        const list = [...this.reviews()];
        const idx = list.findIndex(r => r.id === review.id);
        if (idx >= 0) list[idx] = updated;
        this.reviews.set(list);
        this.replyText = '';
        this.replying.set(false);
        this.loadStats();
      },
      error: () => this.replying.set(false)
    });
  }

  markResolved(review: Review) {
    this.reviewService.updateStatus(review.id, 'resolved').subscribe(() => {
      const list = [...this.reviews()];
      const idx = list.findIndex(r => r.id === review.id);
      if (idx >= 0) list[idx] = { ...list[idx], status: 'resolved' };
      this.reviews.set(list);
      this.loadStats();
    });
  }

  deleteReview(review: Review) {
    if (!confirm('Delete this review?')) return;
    this.reviewService.delete(review.id).subscribe(() => {
      this.reviews.set(this.reviews().filter(r => r.id !== review.id));
      this.loadStats();
    });
  }

  goToPage(page: number) {
    if (page < 1 || page > this.totalPages) return;
    this.currentPage = page;
    this.loadReviews();
  }

  getStars(rating: number): string {
    if (rating <= 0) return '☆☆☆☆☆';
    const full = Math.floor(rating);
    const half = rating - full >= 0.5 ? 1 : 0;
    return '★'.repeat(full) + (half ? '½' : '') + '☆'.repeat(5 - full - half);
  }

  getRatingPercent(rating: number): number {
    const s = this.stats();
    if (!s || s.total_reviews === 0) return 0;
    return ((s.rating_breakdown[rating] || 0) / s.total_reviews) * 100;
  }

  formatDate(dateStr: string): string {
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
