import { Component, inject, OnInit, OnDestroy, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { ImportService } from '../../core/services/import.service';
import { MarketplaceService } from '../../core/services/marketplace.service';
import { ProductService, Product, ProductsResponse } from '../../core/services/product.service';

@Component({
  selector: 'app-products',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="products-page">
      <div class="page-header">
        <div class="header-left">
          <h1>Products</h1>
          <span class="product-count">{{ totalProducts() }} products</span>
        </div>
        <div class="header-actions">
          <div class="search-box">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
            </svg>
            <input type="text" placeholder="Search products..." [(ngModel)]="searchQuery" (input)="onSearch()">
          </div>
          <button class="btn-secondary" (click)="showImportDialog.set(true)">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/>
            </svg>
            Import
          </button>
          <button class="btn-primary" (click)="addProduct()">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
            </svg>
            Add Product
          </button>
        </div>
      </div>

      <!-- Import Progress Banner -->
      @if (importService.isRunning()) {
        <div class="import-banner">
          <div class="import-info">
            <svg class="spinner" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
            </svg>
            <span>Importing products... {{ importService.importStatus().imported }} of {{ importService.importStatus().total }}</span>
          </div>
          <div class="progress-bar">
            <div class="progress-fill" [style.width.%]="importService.progress()"></div>
          </div>
        </div>
      }

      @if (importService.importStatus().status === 'completed') {
        <div class="import-banner success">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/>
          </svg>
          <span>Import completed! {{ importService.importStatus().imported }} products imported.</span>
          <button class="btn-dismiss" (click)="dismissImportBanner()">Dismiss</button>
        </div>
      }

      <!-- Import Dialog -->
      @if (showImportDialog()) {
        <div class="modal-overlay" (click)="showImportDialog.set(false)">
          <div class="modal" (click)="$event.stopPropagation()">
            <div class="modal-header">
              <h2>Import Products</h2>
              <button class="btn-close" (click)="showImportDialog.set(false)">&times;</button>
            </div>
            <div class="modal-body">
              @if (connections().length === 0) {
                <div class="empty-state">
                  <p>No marketplace connections found.</p>
                  <p class="text-muted">Go to Settings to connect a marketplace first.</p>
                </div>
              } @else {
                <p class="modal-description">Select a marketplace connection to import products from:</p>
                <div class="connection-list">
                  @for (conn of connections(); track conn.id) {
                    <div class="connection-option"
                         [class.selected]="selectedConnection() === conn.id"
                         (click)="selectedConnection.set(conn.id)">
                      <div class="conn-info">
                        <span class="conn-marketplace">{{ conn.marketplace | uppercase }}</span>
                        <span class="conn-url">{{ conn.shop_url }}</span>
                      </div>
                      <div class="conn-radio">
                        <div class="radio" [class.active]="selectedConnection() === conn.id"></div>
                      </div>
                    </div>
                  }
                </div>
              }
            </div>
            <div class="modal-footer">
              <button class="btn-secondary" (click)="showImportDialog.set(false)">Cancel</button>
              <button class="btn-primary"
                      [disabled]="!selectedConnection() || importLoading()"
                      (click)="startImport()">
                @if (importLoading()) {
                  Starting...
                } @else {
                  Start Import
                }
              </button>
            </div>
          </div>
        </div>
      }

      <!-- Products Table -->
      @if (loading()) {
        <div class="loading-state">Loading products...</div>
      } @else if (products().length === 0) {
        <div class="empty-state-large">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="var(--text-muted)" stroke-width="1.5">
            <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
            <polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/>
          </svg>
          <h3>No products yet</h3>
          <p>Import products from your connected marketplaces to get started.</p>
          <button class="btn-primary" (click)="showImportDialog.set(true)">Import Products</button>
        </div>
      } @else {
        <div class="products-table-wrapper">
          <table class="products-table">
            <thead>
              <tr>
                <th>Image</th>
                <th>Product</th>
                <th>SKU</th>
                <th>Price</th>
                <th>Stock</th>
                <th>Status</th>
                <th>Marketplace</th>
              </tr>
            </thead>
            <tbody>
              @for (product of products(); track product.id) {
                <tr class="clickable-row" (click)="editProduct(product.id)">
                  <td class="img-cell">
                    @if (product.images && product.images.length > 0) {
                      <img [src]="product.images[0].url" [alt]="product.title" class="product-thumb">
                    } @else {
                      <div class="no-image">
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                          <rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="m21 15-5-5L5 21"/>
                        </svg>
                      </div>
                    }
                  </td>
                  <td>
                    <div class="product-name">{{ product.title }}</div>
                  </td>
                  <td class="sku-cell">{{ product.sku || '\u2014' }}</td>
                  <td class="price-cell">
                    <span class="price">{{ product.base_price | number:'1.2-2' }} \u20B4</span>
                    @if (product.sale_price && product.sale_price !== product.base_price) {
                      <span class="old-price">{{ product.sale_price | number:'1.2-2' }} \u20B4</span>
                    }
                  </td>
                  <td class="stock-cell">
                    <span [class.low-stock]="product.stock < 5 && product.stock > 0"
                          [class.out-of-stock]="product.stock === 0">
                      {{ product.stock }}
                    </span>
                  </td>
                  <td>
                    <span class="status-badge" [class]="'status-' + product.status">
                      {{ product.status }}
                    </span>
                  </td>
                  <td>
                    @for (md of product.marketplace_data; track md.marketplace) {
                      <span class="marketplace-badge">{{ md.marketplace }}</span>
                    }
                  </td>
                </tr>
              }
            </tbody>
          </table>
        </div>

        @if (totalProducts() > pageSize) {
          <div class="pagination">
            <button class="btn-page" [disabled]="currentPage() <= 1" (click)="goToPage(currentPage() - 1)">Previous</button>
            <span class="page-info">Page {{ currentPage() }} of {{ totalPages() }}</span>
            <button class="btn-page" [disabled]="currentPage() >= totalPages()" (click)="goToPage(currentPage() + 1)">Next</button>
          </div>
        }
      }
    </div>
  `,
  styles: [`
    .products-page { padding: 24px; }
    .page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
    .header-left { display: flex; align-items: baseline; gap: 12px; }
    .header-left h1 { font-size: 24px; font-weight: 600; }
    .product-count { color: var(--text-muted); font-size: 14px; }
    .header-actions { display: flex; gap: 12px; align-items: center; }

    .search-box {
      display: flex; align-items: center; gap: 8px;
      background: var(--bg-surface); border: 1px solid var(--border);
      border-radius: 8px; padding: 8px 12px; color: var(--text-secondary);
    }
    .search-box input {
      background: transparent; border: none; outline: none;
      color: var(--text-primary); font-size: 14px; width: 200px;
    }
    .search-box input::placeholder { color: var(--text-muted); }

    .btn-primary {
      display: flex; align-items: center; gap: 8px;
      background: var(--accent-blue); color: white; border: none;
      padding: 8px 16px; border-radius: 8px; font-size: 14px;
      cursor: pointer; font-weight: 500; white-space: nowrap;
    }
    .btn-primary:hover { background: var(--accent-blue-hover); }
    .btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }

    .btn-secondary {
      background: var(--bg-surface-light); color: var(--text-primary);
      border: 1px solid var(--border); padding: 8px 16px;
      border-radius: 8px; font-size: 14px; cursor: pointer;
    }
    .btn-secondary:hover { background: var(--border); }

    .import-banner {
      background: var(--bg-surface); border: 1px solid var(--border);
      border-radius: 8px; padding: 16px; margin-bottom: 24px;
    }
    .import-banner.success {
      border-color: var(--accent-green); display: flex; align-items: center; gap: 12px;
    }
    .import-banner.success svg { color: var(--accent-green); }
    .import-info { display: flex; align-items: center; gap: 12px; margin-bottom: 8px; }

    .spinner { animation: spin 1s linear infinite; }
    @keyframes spin { to { transform: rotate(360deg); } }

    .progress-bar {
      height: 4px; background: var(--bg-surface-light); border-radius: 2px; overflow: hidden;
    }
    .progress-fill {
      height: 100%; background: var(--accent-blue); border-radius: 2px;
      transition: width 0.3s ease;
    }
    .btn-dismiss {
      margin-left: auto; background: transparent; border: 1px solid var(--border);
      color: var(--text-secondary); padding: 4px 12px; border-radius: 6px;
      cursor: pointer; font-size: 13px;
    }

    .modal-overlay {
      position: fixed; inset: 0; background: rgba(0,0,0,0.6);
      display: flex; align-items: center; justify-content: center; z-index: 1000;
    }
    .modal {
      background: var(--bg-surface); border: 1px solid var(--border);
      border-radius: 12px; width: 480px; max-width: 90vw;
    }
    .modal-header {
      display: flex; justify-content: space-between; align-items: center;
      padding: 20px 24px; border-bottom: 1px solid var(--border);
    }
    .modal-header h2 { font-size: 18px; font-weight: 600; }
    .btn-close {
      background: transparent; border: none; color: var(--text-muted);
      font-size: 24px; cursor: pointer; padding: 0; line-height: 1;
    }
    .modal-body { padding: 24px; }
    .modal-description { color: var(--text-secondary); margin-bottom: 16px; font-size: 14px; }
    .modal-footer {
      display: flex; justify-content: flex-end; gap: 12px;
      padding: 16px 24px; border-top: 1px solid var(--border);
    }

    .connection-list { display: flex; flex-direction: column; gap: 8px; }
    .connection-option {
      display: flex; justify-content: space-between; align-items: center;
      padding: 12px 16px; background: var(--bg-surface-light);
      border: 1px solid var(--border); border-radius: 8px; cursor: pointer;
    }
    .connection-option:hover { border-color: var(--accent-blue); }
    .connection-option.selected { border-color: var(--accent-blue); background: rgba(59, 130, 246, 0.1); }
    .conn-info { display: flex; flex-direction: column; gap: 4px; }
    .conn-marketplace { font-weight: 600; font-size: 14px; }
    .conn-url { color: var(--text-muted); font-size: 13px; }
    .radio {
      width: 18px; height: 18px; border-radius: 50%;
      border: 2px solid var(--border);
    }
    .radio.active {
      border-color: var(--accent-blue);
      background: var(--accent-blue);
      box-shadow: inset 0 0 0 3px var(--bg-surface);
    }

    .empty-state { text-align: center; padding: 24px; color: var(--text-secondary); }
    .text-muted { color: var(--text-muted); font-size: 13px; }
    .loading-state { text-align: center; padding: 48px; color: var(--text-muted); }
    .empty-state-large {
      text-align: center; padding: 64px 24px;
      display: flex; flex-direction: column; align-items: center; gap: 12px;
    }
    .empty-state-large h3 { font-size: 18px; font-weight: 600; }
    .empty-state-large p { color: var(--text-muted); margin-bottom: 8px; }

    .products-table-wrapper { overflow-x: auto; }
    .products-table {
      width: 100%; border-collapse: collapse;
      background: var(--bg-surface); border-radius: 8px; overflow: hidden;
    }
    .products-table th {
      text-align: left; padding: 12px 16px; font-size: 12px;
      font-weight: 600; color: var(--text-muted); text-transform: uppercase;
      letter-spacing: 0.05em; border-bottom: 1px solid var(--border);
      background: var(--bg-surface-light);
    }
    .products-table td {
      padding: 12px 16px; border-bottom: 1px solid var(--border);
      font-size: 14px; vertical-align: middle;
    }
    .products-table tr:hover { background: var(--bg-surface-light); }
    .clickable-row { cursor: pointer; }

    .img-cell { width: 60px; }
    .product-thumb {
      width: 44px; height: 44px; border-radius: 6px; object-fit: cover;
      background: var(--bg-surface-light);
    }
    .no-image {
      width: 44px; height: 44px; border-radius: 6px;
      background: var(--bg-surface-light); display: flex;
      align-items: center; justify-content: center; color: var(--text-muted);
    }
    .product-name { font-weight: 500; max-width: 300px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .sku-cell { color: var(--text-muted); font-family: monospace; font-size: 13px; }
    .price-cell { white-space: nowrap; }
    .price { font-weight: 600; }
    .old-price { color: var(--text-muted); text-decoration: line-through; font-size: 12px; margin-left: 6px; }
    .low-stock { color: var(--warning); }
    .out-of-stock { color: var(--error); }

    .status-badge {
      padding: 2px 8px; border-radius: 4px; font-size: 12px; font-weight: 500;
    }
    .status-active { background: rgba(16, 185, 129, 0.15); color: var(--accent-green); }
    .status-draft { background: rgba(148, 163, 184, 0.15); color: var(--text-secondary); }
    .status-archived { background: rgba(100, 116, 139, 0.15); color: var(--text-muted); }
    .marketplace-badge {
      padding: 2px 8px; border-radius: 4px; font-size: 11px;
      font-weight: 600; text-transform: uppercase;
      background: rgba(59, 130, 246, 0.15); color: var(--accent-blue);
    }

    .pagination {
      display: flex; justify-content: center; align-items: center; gap: 16px;
      padding: 24px;
    }
    .btn-page {
      background: var(--bg-surface); border: 1px solid var(--border);
      color: var(--text-primary); padding: 6px 16px; border-radius: 6px;
      cursor: pointer; font-size: 13px;
    }
    .btn-page:disabled { opacity: 0.4; cursor: not-allowed; }
    .btn-page:hover:not(:disabled) { background: var(--bg-surface-light); }
    .page-info { color: var(--text-muted); font-size: 13px; }
  `]
})
export class ProductsComponent implements OnInit, OnDestroy {
  private productService = inject(ProductService);
  private router = inject(Router);
  importService = inject(ImportService);
  private marketplaceService = inject(MarketplaceService);

  products = signal<Product[]>([]);
  totalProducts = signal(0);
  currentPage = signal(1);
  pageSize = 20;
  totalPages = signal(1);
  loading = signal(true);
  searchQuery = '';

  showImportDialog = signal(false);
  connections = signal<any[]>([]);
  selectedConnection = signal<string | null>(null);
  importLoading = signal(false);

  private searchTimeout: any;
  private productPollInterval: any;

  ngOnInit() {
    this.loadProducts();
    this.loadConnections();
    this.importService.startPolling();
  }

  ngOnDestroy() {
    this.importService.stopPolling();
    if (this.productPollInterval) {
      clearInterval(this.productPollInterval);
    }
  }

  loadProducts() {
    this.loading.set(true);
    this.productService.list({
      page: this.currentPage(),
      limit: this.pageSize,
      search: this.searchQuery || undefined,
    }).subscribe({
      next: (res) => {
        this.products.set(res.products || []);
        this.totalProducts.set(res.total);
        this.totalPages.set(Math.ceil(res.total / this.pageSize));
        this.loading.set(false);
      },
      error: () => {
        this.products.set([]);
        this.loading.set(false);
      }
    });
  }

  loadConnections() {
    this.marketplaceService.list().subscribe({
      next: (conns: any[]) => this.connections.set(conns),
    });
  }

  onSearch() {
    clearTimeout(this.searchTimeout);
    this.searchTimeout = setTimeout(() => {
      this.currentPage.set(1);
      this.loadProducts();
    }, 300);
  }

  goToPage(page: number) {
    this.currentPage.set(page);
    this.loadProducts();
  }

  startImport() {
    const connId = this.selectedConnection();
    if (!connId) return;

    this.importLoading.set(true);
    this.importService.startImport(connId).subscribe({
      next: () => {
        this.importLoading.set(false);
        this.showImportDialog.set(false);
        this.importService.startPolling();
        this.productPollInterval = setInterval(() => {
          this.loadProducts();
          if (!this.importService.isRunning()) {
            clearInterval(this.productPollInterval);
            this.productPollInterval = null;
            this.loadProducts();
          }
        }, 5000);
      },
      error: () => {
        this.importLoading.set(false);
      }
    });
  }

  addProduct() {
    this.router.navigate(['/products', 'new']);
  }

  editProduct(id: string) {
    this.router.navigate(['/products', id]);
  }

  dismissImportBanner() {
    this.importService.importStatus.set({
      total: 0, imported: 0, skipped: 0, failed: 0, status: 'idle'
    });
    this.loadProducts();
  }
}
