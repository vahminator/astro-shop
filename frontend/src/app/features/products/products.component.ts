import { Component } from '@angular/core';

@Component({
  selector: 'app-products',
  standalone: true,
  template: `
    <div class="products-page">
      <div class="page-header">
        <h2>Product Catalog</h2>
        <button class="btn-primary">+ Add Product</button>
      </div>
      <div class="empty-state">
        <p>No products yet. Import from a marketplace or add manually.</p>
      </div>
    </div>
  `,
  styles: [`
    .page-header {
      display: flex; justify-content: space-between; align-items: center;
      margin-bottom: 24px;
    }
    .page-header h2 { margin: 0; font-size: 20px; font-weight: 600; }
    .btn-primary {
      padding: 8px 20px; background: #3b82f6; color: white; border: none;
      border-radius: 8px; font-size: 14px; font-weight: 500; cursor: pointer;
    }
    .btn-primary:hover { background: #2563eb; }
    .empty-state {
      text-align: center; padding: 60px 20px;
      background: #111d32; border: 1px dashed #1e3a5f;
      border-radius: 10px; color: #64748b;
    }
  `]
})
export class ProductsComponent {}
