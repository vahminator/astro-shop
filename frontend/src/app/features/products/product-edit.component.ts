import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { Product, ProductService } from '../../core/services/product.service';

@Component({
  selector: 'app-product-edit',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="product-edit-page">
      <div class="page-header">
        <button class="btn-back" (click)="goBack()">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="m15 18-6-6 6-6"/>
          </svg>
          Back to Products
        </button>
        <h1>{{ isNew ? 'New Product' : 'Edit Product' }}</h1>
      </div>

      @if (loading()) {
        <div class="loading-state">Loading product...</div>
      } @else {
        <form (ngSubmit)="save()" class="product-form">
          <div class="form-grid">
            <!-- Left column: Main info -->
            <div class="form-column main-column">
              <div class="card">
                <h3 class="card-title">General Information</h3>
                <div class="form-group">
                  <label>Title <span class="required">*</span></label>
                  <input type="text" [(ngModel)]="form.title" name="title" required placeholder="Product title">
                </div>
                <div class="form-group">
                  <label>Description</label>
                  <textarea [(ngModel)]="form.description" name="description" rows="5" placeholder="Product description..."></textarea>
                </div>
                <div class="form-row">
                  <div class="form-group">
                    <label>SKU</label>
                    <input type="text" [(ngModel)]="form.sku" name="sku" placeholder="SKU-001">
                  </div>
                  <div class="form-group">
                    <label>EAN</label>
                    <input type="text" [(ngModel)]="form.ean" name="ean" placeholder="Barcode">
                  </div>
                </div>
              </div>

              <div class="card">
                <h3 class="card-title">Pricing & Stock</h3>
                <div class="form-row triple">
                  <div class="form-group">
                    <label>Base Price (₴)</label>
                    <input type="number" [(ngModel)]="form.base_price" name="base_price" min="0" step="0.01" placeholder="0.00">
                  </div>
                  <div class="form-group">
                    <label>Sale Price (₴)</label>
                    <input type="number" [(ngModel)]="form.sale_price" name="sale_price" min="0" step="0.01" placeholder="Optional">
                  </div>
                  <div class="form-group">
                    <label>Stock</label>
                    <input type="number" [(ngModel)]="form.stock" name="stock" min="0" step="1" placeholder="0">
                  </div>
                </div>
              </div>

              <!-- Images -->
              <div class="card">
                <div class="card-header-row">
                  <h3 class="card-title">Images</h3>
                  @if (!isNew) {
                    <label class="btn-upload" for="imageUpload">
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/>
                      </svg>
                      Upload
                    </label>
                    <input type="file" id="imageUpload" accept="image/*" multiple (change)="onImageSelect($event)" style="display:none">
                  }
                </div>

                @if (isNew) {
                  <p class="text-muted">Save the product first before uploading images.</p>
                } @else if (images().length === 0 && !uploadingImage()) {
                  <div class="empty-images">
                    <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="var(--text-muted)" stroke-width="1.5">
                      <rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="m21 15-5-5L5 21"/>
                    </svg>
                    <p>No images yet</p>
                  </div>
                } @else {
                  <div class="image-gallery">
                    @for (img of images(); track img.id) {
                      <div class="image-item">
                        <img [src]="img.url" [alt]="'Image ' + img.position">
                        <button class="btn-remove-image" (click)="removeImage(img.id)" title="Remove image">
                          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
                          </svg>
                        </button>
                      </div>
                    }
                    @if (uploadingImage()) {
                      <div class="image-item uploading">
                        <div class="upload-spinner"></div>
                      </div>
                    }
                  </div>
                }
              </div>

              <!-- Attributes -->
              <div class="card">
                <div class="card-header-row">
                  <h3 class="card-title">Attributes</h3>
                  <button type="button" class="btn-add" (click)="addAttribute()">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
                    </svg>
                    Add
                  </button>
                </div>

                @if (form.attributes.length === 0) {
                  <p class="text-muted">No attributes. Add custom attributes like color, size, material.</p>
                } @else {
                  <div class="attributes-list">
                    @for (attr of form.attributes; track $index) {
                      <div class="attribute-row">
                        <input type="text" [(ngModel)]="attr.name" [name]="'attr_name_' + $index" placeholder="Name (e.g. Color)">
                        <input type="text" [(ngModel)]="attr.value" [name]="'attr_value_' + $index" placeholder="Value (e.g. Red)">
                        <button type="button" class="btn-remove-attr" (click)="removeAttribute($index)">
                          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
                          </svg>
                        </button>
                      </div>
                    }
                  </div>
                }
              </div>
            </div>

            <!-- Right column: Status, marketplace data -->
            <div class="form-column side-column">
              <div class="card">
                <h3 class="card-title">Status</h3>
                <div class="form-group">
                  <select [(ngModel)]="form.status" name="status">
                    <option value="draft">Draft</option>
                    <option value="active">Active</option>
                    <option value="archived">Archived</option>
                  </select>
                </div>
              </div>

              @if (!isNew && marketplaceData().length > 0) {
                <div class="card">
                  <h3 class="card-title">Marketplace Data</h3>
                  @for (md of marketplaceData(); track md.marketplace) {
                    <div class="mp-item">
                      <div class="mp-header">
                        <span class="marketplace-badge">{{ md.marketplace }}</span>
                        <span class="mp-status" [class]="'status-' + md.status">{{ md.status }}</span>
                      </div>
                      @if (md.external_id) {
                        <div class="mp-detail">
                          <span class="mp-label">External ID:</span>
                          <span class="mp-value">{{ md.external_id }}</span>
                        </div>
                      }
                      @if (md.marketplace_url) {
                        <a [href]="md.marketplace_url" target="_blank" class="mp-link">View on Marketplace →</a>
                      }
                    </div>
                  }
                </div>
              }

              @if (!isNew) {
                <div class="card danger-zone">
                  <h3 class="card-title">Danger Zone</h3>
                  <p class="text-muted">Permanently delete this product and all its data.</p>
                  <button type="button" class="btn-danger" (click)="deleteProduct()">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
                    </svg>
                    Delete Product
                  </button>
                </div>
              }
            </div>
          </div>

          <!-- Footer actions -->
          <div class="form-actions">
            <button type="button" class="btn-secondary" (click)="goBack()">Cancel</button>
            <button type="submit" class="btn-primary" [disabled]="saving()">
              @if (saving()) {
                Saving...
              } @else {
                {{ isNew ? 'Create Product' : 'Save Changes' }}
              }
            </button>
          </div>
        </form>
      }
    </div>
  `,
  styles: [`
    .product-edit-page { padding: 24px; max-width: 1200px; }

    .page-header { margin-bottom: 24px; }
    .btn-back {
      display: inline-flex; align-items: center; gap: 6px;
      background: transparent; border: none; color: var(--text-muted);
      cursor: pointer; font-size: 13px; padding: 0; margin-bottom: 12px;
    }
    .btn-back:hover { color: var(--text-primary); }
    .page-header h1 { font-size: 24px; font-weight: 600; }

    .loading-state { text-align: center; padding: 48px; color: var(--text-muted); }

    .form-grid {
      display: grid; grid-template-columns: 1fr 320px; gap: 24px;
    }
    @media (max-width: 900px) {
      .form-grid { grid-template-columns: 1fr; }
    }

    .form-column { display: flex; flex-direction: column; gap: 20px; }

    .card {
      background: var(--bg-surface); border: 1px solid var(--border);
      border-radius: 10px; padding: 20px;
    }
    .card-title { font-size: 15px; font-weight: 600; margin-bottom: 16px; }
    .card-header-row {
      display: flex; justify-content: space-between; align-items: center;
      margin-bottom: 16px;
    }
    .card-header-row .card-title { margin-bottom: 0; }

    .form-group { margin-bottom: 14px; }
    .form-group:last-child { margin-bottom: 0; }
    .form-group label {
      display: block; font-size: 13px; font-weight: 500;
      color: var(--text-secondary); margin-bottom: 6px;
    }
    .required { color: var(--error); }

    .form-group input,
    .form-group textarea,
    .form-group select {
      width: 100%; padding: 9px 12px;
      background: var(--bg-primary); border: 1px solid var(--border);
      border-radius: 8px; color: var(--text-primary); font-size: 14px;
      outline: none; font-family: inherit;
    }
    .form-group input:focus,
    .form-group textarea:focus,
    .form-group select:focus {
      border-color: var(--accent-blue);
    }
    .form-group textarea { resize: vertical; min-height: 100px; }
    .form-group select { cursor: pointer; }
    .form-group input::placeholder,
    .form-group textarea::placeholder { color: var(--text-muted); }

    .form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
    .form-row.triple { grid-template-columns: 1fr 1fr 1fr; }

    /* Images */
    .btn-upload {
      display: inline-flex; align-items: center; gap: 6px;
      background: var(--bg-surface-light); border: 1px solid var(--border);
      color: var(--text-primary); padding: 5px 12px; border-radius: 6px;
      font-size: 13px; cursor: pointer;
    }
    .btn-upload:hover { background: var(--border); }

    .empty-images {
      display: flex; flex-direction: column; align-items: center;
      gap: 8px; padding: 24px; color: var(--text-muted); font-size: 14px;
    }

    .image-gallery {
      display: grid; grid-template-columns: repeat(auto-fill, minmax(100px, 1fr)); gap: 10px;
    }
    .image-item {
      position: relative; aspect-ratio: 1; border-radius: 8px; overflow: hidden;
      background: var(--bg-surface-light);
    }
    .image-item img {
      width: 100%; height: 100%; object-fit: cover;
    }
    .btn-remove-image {
      position: absolute; top: 4px; right: 4px;
      background: rgba(0,0,0,0.7); border: none; color: white;
      width: 24px; height: 24px; border-radius: 50%;
      display: flex; align-items: center; justify-content: center;
      cursor: pointer; opacity: 0; transition: opacity 0.2s;
    }
    .image-item:hover .btn-remove-image { opacity: 1; }

    .image-item.uploading {
      display: flex; align-items: center; justify-content: center;
    }
    .upload-spinner {
      width: 24px; height: 24px; border: 2px solid var(--border);
      border-top-color: var(--accent-blue); border-radius: 50%;
      animation: spin 0.8s linear infinite;
    }
    @keyframes spin { to { transform: rotate(360deg); } }

    /* Attributes */
    .btn-add {
      display: inline-flex; align-items: center; gap: 6px;
      background: var(--bg-surface-light); border: 1px solid var(--border);
      color: var(--text-primary); padding: 5px 12px; border-radius: 6px;
      font-size: 13px; cursor: pointer;
    }
    .btn-add:hover { background: var(--border); }

    .attributes-list { display: flex; flex-direction: column; gap: 8px; }
    .attribute-row {
      display: grid; grid-template-columns: 1fr 1fr 32px; gap: 8px;
      align-items: center;
    }
    .attribute-row input {
      padding: 8px 10px; background: var(--bg-primary);
      border: 1px solid var(--border); border-radius: 6px;
      color: var(--text-primary); font-size: 13px; outline: none;
    }
    .attribute-row input:focus { border-color: var(--accent-blue); }
    .attribute-row input::placeholder { color: var(--text-muted); }
    .btn-remove-attr {
      background: transparent; border: none; color: var(--text-muted);
      cursor: pointer; padding: 4px; display: flex; align-items: center;
      justify-content: center;
    }
    .btn-remove-attr:hover { color: var(--error); }

    /* Marketplace data */
    .mp-item {
      padding: 12px; background: var(--bg-primary); border-radius: 8px;
      margin-bottom: 8px;
    }
    .mp-item:last-child { margin-bottom: 0; }
    .mp-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
    .marketplace-badge {
      padding: 2px 8px; border-radius: 4px; font-size: 11px;
      font-weight: 600; text-transform: uppercase;
      background: rgba(59,130,246,0.15); color: var(--accent-blue);
    }
    .mp-status { font-size: 12px; font-weight: 500; }
    .status-published { color: var(--accent-green); }
    .status-pending { color: var(--warning); }
    .status-error { color: var(--error); }
    .mp-detail { font-size: 13px; color: var(--text-muted); margin-bottom: 4px; }
    .mp-label { margin-right: 6px; }
    .mp-value { color: var(--text-secondary); font-family: monospace; }
    .mp-link {
      font-size: 13px; color: var(--accent-blue); text-decoration: none;
    }
    .mp-link:hover { text-decoration: underline; }

    /* Danger zone */
    .danger-zone { border-color: rgba(239,68,68,0.3); }
    .btn-danger {
      display: inline-flex; align-items: center; gap: 8px;
      background: rgba(239,68,68,0.15); color: var(--error);
      border: 1px solid rgba(239,68,68,0.3); padding: 8px 16px;
      border-radius: 8px; font-size: 13px; cursor: pointer; margin-top: 8px;
    }
    .btn-danger:hover { background: rgba(239,68,68,0.25); }

    .text-muted { color: var(--text-muted); font-size: 13px; }

    /* Form actions */
    .form-actions {
      display: flex; justify-content: flex-end; gap: 12px;
      margin-top: 24px; padding-top: 20px;
      border-top: 1px solid var(--border);
    }
    .btn-primary {
      display: flex; align-items: center; gap: 8px;
      background: var(--accent-blue); color: white; border: none;
      padding: 10px 24px; border-radius: 8px; font-size: 14px;
      cursor: pointer; font-weight: 500;
    }
    .btn-primary:hover { background: var(--accent-blue-hover); }
    .btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
    .btn-secondary {
      background: var(--bg-surface); color: var(--text-primary);
      border: 1px solid var(--border); padding: 10px 24px;
      border-radius: 8px; font-size: 14px; cursor: pointer;
    }
    .btn-secondary:hover { background: var(--bg-surface-light); }
  `]
})
export class ProductEditComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private productService = inject(ProductService);

  isNew = true;
  productId = '';
  loading = signal(false);
  saving = signal(false);
  uploadingImage = signal(false);
  images = signal<{ id: string; url: string; position: number }[]>([]);
  marketplaceData = signal<{ marketplace: string; external_id: string; status: string; marketplace_url: string }[]>([]);

  form = {
    title: '',
    description: '',
    sku: '',
    ean: '',
    base_price: 0,
    sale_price: null as number | null,
    stock: 0,
    status: 'draft',
    attributes: [] as { name: string; value: string }[],
  };

  ngOnInit() {
    const id = this.route.snapshot.paramMap.get('id');
    if (id && id !== 'new') {
      this.isNew = false;
      this.productId = id;
      this.loadProduct();
    }
  }

  loadProduct() {
    this.loading.set(true);
    this.productService.get(this.productId).subscribe({
      next: (p) => {
        this.form.title = p.title;
        this.form.description = p.description || '';
        this.form.sku = p.sku || '';
        this.form.ean = p.ean || '';
        this.form.base_price = p.base_price;
        this.form.sale_price = p.sale_price;
        this.form.stock = p.stock;
        this.form.status = p.status;
        this.form.attributes = (p.attributes || []).map(a => ({ name: a.name, value: a.value }));
        this.images.set(p.images || []);
        this.marketplaceData.set(p.marketplace_data || []);
        this.loading.set(false);
      },
      error: () => {
        this.loading.set(false);
        this.router.navigate(['/products']);
      }
    });
  }

  save() {
    if (!this.form.title.trim()) return;
    this.saving.set(true);

    const payload: any = {
      title: this.form.title,
      description: this.form.description,
      sku: this.form.sku,
      ean: this.form.ean,
      base_price: this.form.base_price || 0,
      stock: this.form.stock || 0,
      status: this.form.status,
      attributes: this.form.attributes.filter(a => a.name.trim() && a.value.trim()),
    };

    if (this.form.sale_price !== null && this.form.sale_price !== undefined) {
      payload.sale_price = this.form.sale_price;
    }

    if (this.isNew) {
      this.productService.create(payload).subscribe({
        next: (p) => {
          this.saving.set(false);
          this.router.navigate(['/products', p.id]);
        },
        error: () => this.saving.set(false),
      });
    } else {
      this.productService.update(this.productId, payload).subscribe({
        next: (p) => {
          this.saving.set(false);
          this.images.set(p.images || []);
          this.marketplaceData.set(p.marketplace_data || []);
        },
        error: () => this.saving.set(false),
      });
    }
  }

  onImageSelect(event: Event) {
    const input = event.target as HTMLInputElement;
    const files = input.files;
    if (!files || files.length === 0) return;

    for (let i = 0; i < files.length; i++) {
      this.uploadSingleImage(files[i]);
    }
    input.value = '';
  }

  private uploadSingleImage(file: File) {
    this.uploadingImage.set(true);
    this.productService.uploadImage(this.productId, file).subscribe({
      next: (img) => {
        this.images.update(imgs => [...imgs, img]);
        this.uploadingImage.set(false);
      },
      error: () => this.uploadingImage.set(false),
    });
  }

  removeImage(imageId: string) {
    this.productService.deleteImage(this.productId, imageId).subscribe({
      next: () => {
        this.images.update(imgs => imgs.filter(i => i.id !== imageId));
      },
    });
  }

  addAttribute() {
    this.form.attributes.push({ name: '', value: '' });
  }

  removeAttribute(index: number) {
    this.form.attributes.splice(index, 1);
  }

  deleteProduct() {
    if (!confirm('Are you sure you want to delete this product? This action cannot be undone.')) return;
    this.productService.delete(this.productId).subscribe({
      next: () => this.router.navigate(['/products']),
    });
  }

  goBack() {
    this.router.navigate(['/products']);
  }
}
