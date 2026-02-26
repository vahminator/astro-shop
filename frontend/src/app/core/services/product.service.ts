import { Injectable } from '@angular/core';
import { ApiService } from './api.service';
import { Observable } from 'rxjs';

export interface ProductImage {
  id: string;
  url: string;
  position: number;
}

export interface ProductAttribute {
  name: string;
  value: string;
}

export interface ProductMarketplaceData {
  marketplace: string;
  external_id: string;
  status: string;
  marketplace_url: string;
}

export interface Product {
  id: string;
  title: string;
  description: string;
  sku: string;
  ean: string;
  base_price: number;
  sale_price: number | null;
  stock: number;
  status: string;
  category_id: string | null;
  images: ProductImage[];
  attributes: ProductAttribute[];
  marketplace_data: ProductMarketplaceData[];
  created_at: string;
  updated_at: string;
}

export interface ProductsResponse {
  products: Product[];
  total: number;
  page: number;
  limit: number;
}

export interface CreateProductRequest {
  title: string;
  description?: string;
  sku?: string;
  ean?: string;
  base_price?: number;
  sale_price?: number | null;
  stock?: number;
  status?: string;
  category_id?: string | null;
  attributes?: { name: string; value: string }[];
}

export interface UpdateProductRequest {
  title?: string;
  description?: string;
  sku?: string;
  ean?: string;
  base_price?: number;
  sale_price?: number | null;
  stock?: number;
  status?: string;
  category_id?: string | null;
  attributes?: { name: string; value: string }[];
}

@Injectable({ providedIn: 'root' })
export class ProductService {
  constructor(private api: ApiService) {}

  list(params?: { page?: number; limit?: number; search?: string; status?: string }): Observable<ProductsResponse> {
    const query = new URLSearchParams();
    if (params?.page) query.set('page', params.page.toString());
    if (params?.limit) query.set('limit', params.limit.toString());
    if (params?.search) query.set('search', params.search);
    if (params?.status) query.set('status', params.status);
    return this.api.get<ProductsResponse>(`/products?${query.toString()}`);
  }

  get(id: string): Observable<Product> {
    return this.api.get<Product>(`/products/${id}`);
  }

  create(req: CreateProductRequest): Observable<Product> {
    return this.api.post<Product>('/products', req);
  }

  update(id: string, req: UpdateProductRequest): Observable<Product> {
    return this.api.put<Product>(`/products/${id}`, req);
  }

  delete(id: string): Observable<void> {
    return this.api.delete<void>(`/products/${id}`);
  }

  uploadImage(productId: string, file: File): Observable<ProductImage> {
    const formData = new FormData();
    formData.append('image', file);
    return this.api.postFormData<ProductImage>(`/products/${productId}/images`, formData);
  }

  deleteImage(productId: string, imageId: string): Observable<void> {
    return this.api.delete<void>(`/products/${productId}/images/${imageId}`);
  }
}
