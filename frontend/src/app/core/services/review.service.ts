import { Injectable } from '@angular/core';
import { ApiService } from './api.service';
import { Observable } from 'rxjs';

export interface Review {
  id: string;
  marketplace: string;
  external_id: string;
  product_id: string | null;
  product_title: string;
  author_name: string;
  rating: number;
  body: string;
  status: string;
  responses: ReviewResponseItem[];
  created_at: string;
}

export interface ReviewResponseItem {
  id: string;
  body: string;
  created_at: string;
}

export interface ReviewsResponse {
  reviews: Review[];
  total: number;
  page: number;
  limit: number;
}

export interface ReviewStats {
  total_reviews: number;
  average_rating: number;
  rating_breakdown: { [key: number]: number };
  new_count: number;
  responded_count: number;
  resolved_count: number;
}

@Injectable({ providedIn: 'root' })
export class ReviewService {
  constructor(private api: ApiService) {}

  list(params?: { page?: number; limit?: number; status?: string; rating?: number; marketplace?: string; search?: string }): Observable<ReviewsResponse> {
    const query = new URLSearchParams();
    if (params?.page) query.set('page', params.page.toString());
    if (params?.limit) query.set('limit', params.limit.toString());
    if (params?.status) query.set('status', params.status);
    if (params?.rating) query.set('rating', params.rating.toString());
    if (params?.marketplace) query.set('marketplace', params.marketplace);
    if (params?.search) query.set('search', params.search);
    const qs = query.toString();
    return this.api.get<ReviewsResponse>(`/reviews${qs ? '?' + qs : ''}`);
  }

  get(id: string): Observable<Review> {
    return this.api.get<Review>(`/reviews/${id}`);
  }

  create(data: { marketplace: string; author_name: string; rating: number; body?: string; product_id?: string }): Observable<Review> {
    return this.api.post<Review>('/reviews', data);
  }

  reply(reviewId: string, body: string): Observable<Review> {
    return this.api.post<Review>(`/reviews/${reviewId}/reply`, { body });
  }

  updateStatus(reviewId: string, status: string): Observable<any> {
    return this.api.put(`/reviews/${reviewId}/status`, { status });
  }

  delete(reviewId: string): Observable<void> {
    return this.api.delete<void>(`/reviews/${reviewId}`);
  }

  sync(marketplaceId: string): Observable<any> {
    return this.api.post('/reviews/sync', { marketplace_id: marketplaceId });
  }

  stats(): Observable<ReviewStats> {
    return this.api.get<ReviewStats>('/reviews/stats');
  }
}
