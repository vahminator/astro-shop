import { Injectable } from '@angular/core';
import { ApiService } from './api.service';
import { Observable } from 'rxjs';

export interface Conversation {
  id: string;
  marketplace: string;
  external_id: string;
  customer_name: string;
  customer_id: string;
  subject: string;
  status: string;
  last_message_at: string | null;
  last_message: string | null;
  unread_count: number;
  created_at: string;
}

export interface Message {
  id: string;
  external_id: string;
  sender: string;
  body: string;
  is_read: boolean;
  created_at: string;
}

export interface ConversationDetail extends Conversation {
  messages: Message[];
}

export interface MessageTemplate {
  id: string;
  title: string;
  body: string;
  shortcut: string;
}

@Injectable({ providedIn: 'root' })
export class InboxService {
  constructor(private api: ApiService) {}

  listConversations(status?: string): Observable<Conversation[]> {
    const query = status ? `?status=${status}` : '';
    return this.api.get<Conversation[]>(`/inbox${query}`);
  }

  getConversation(id: string): Observable<ConversationDetail> {
    return this.api.get<ConversationDetail>(`/inbox/${id}`);
  }

  sendMessage(conversationId: string, body: string): Observable<any> {
    return this.api.post(`/inbox/${conversationId}/messages`, { body });
  }

  updateStatus(conversationId: string, status: string): Observable<any> {
    return this.api.put(`/inbox/${conversationId}/status`, { status });
  }

  syncInbox(marketplaceId: string): Observable<any> {
    return this.api.post('/inbox/sync', { marketplace_id: marketplaceId });
  }

  listTemplates(): Observable<MessageTemplate[]> {
    return this.api.get<MessageTemplate[]>('/templates');
  }

  createTemplate(title: string, body: string, shortcut?: string): Observable<MessageTemplate> {
    return this.api.post<MessageTemplate>('/templates', { title, body, shortcut });
  }

  deleteTemplate(id: string): Observable<void> {
    return this.api.delete<void>(`/templates/${id}`);
  }
}
