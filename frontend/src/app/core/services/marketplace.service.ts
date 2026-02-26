import { Injectable } from '@angular/core';
import { ApiService } from './api.service';
import { Observable } from 'rxjs';

export interface MarketplaceConnection {
  id: string;
  marketplace: string;
  shop_url: string;
  shop_name: string;
  is_active: boolean;
  last_sync_at?: string;
  created_at: string;
}

export interface ConnectRequest {
  marketplace: string;
  api_key: string;
  shop_url?: string;
}

export interface TestConnectionResponse {
  success: boolean;
  message: string;
}

@Injectable({ providedIn: 'root' })
export class MarketplaceService {
  constructor(private api: ApiService) {}

  list(): Observable<MarketplaceConnection[]> {
    return this.api.get<MarketplaceConnection[]>('/marketplaces');
  }

  connect(req: ConnectRequest): Observable<MarketplaceConnection> {
    return this.api.post<MarketplaceConnection>('/marketplaces', req);
  }

  update(id: string, data: Partial<ConnectRequest & { is_active: boolean }>): Observable<MarketplaceConnection> {
    return this.api.put<MarketplaceConnection>(`/marketplaces/${id}`, data);
  }

  delete(id: string): Observable<void> {
    return this.api.delete<void>(`/marketplaces/${id}`);
  }

  testConnection(id: string): Observable<TestConnectionResponse> {
    return this.api.post<TestConnectionResponse>(`/marketplaces/${id}/test`, {});
  }
}
