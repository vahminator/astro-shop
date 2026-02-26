import { Component, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../core/services/auth.service';

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink],
  template: `
    <div class="auth-container">
      <div class="auth-card">
        <div class="auth-header">
          <span class="auth-logo">&#9889;</span>
          <h1 class="auth-title">Sellflow</h1>
          <p class="auth-subtitle">Create your account</p>
        </div>
        <form (ngSubmit)="onSubmit()" class="auth-form">
          @if (error()) {
            <div class="error-message">{{ error() }}</div>
          }
          <div class="form-group">
            <label for="name">Full name</label>
            <input id="name" type="text" [(ngModel)]="name" name="name"
                   placeholder="John Doe" required>
          </div>
          <div class="form-group">
            <label for="email">Email</label>
            <input id="email" type="email" [(ngModel)]="email" name="email"
                   placeholder="you@example.com" required autocomplete="email">
          </div>
          <div class="form-group">
            <label for="password">Password</label>
            <input id="password" type="password" [(ngModel)]="password" name="password"
                   placeholder="Min 6 characters" required minlength="6">
          </div>
          <button type="submit" class="btn-primary" [disabled]="loading()">
            {{ loading() ? 'Creating account...' : 'Create account' }}
          </button>
        </form>
        <p class="auth-footer">
          Already have an account? <a routerLink="/login">Sign in</a>
        </p>
      </div>
    </div>
  `,
  styles: [`
    .auth-container {
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
      background: #0a1628;
      padding: 20px;
    }
    .auth-card {
      width: 100%;
      max-width: 400px;
      background: #111d32;
      border: 1px solid #1e3a5f;
      border-radius: 12px;
      padding: 40px;
    }
    .auth-header { text-align: center; margin-bottom: 32px; }
    .auth-logo { font-size: 40px; display: block; margin-bottom: 12px; }
    .auth-title {
      font-size: 28px; font-weight: 700; margin: 0 0 8px;
      background: linear-gradient(135deg, #3b82f6, #10b981);
      -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text;
    }
    .auth-subtitle { color: #64748b; margin: 0; font-size: 14px; }
    .auth-form { display: flex; flex-direction: column; gap: 20px; }
    .form-group { display: flex; flex-direction: column; gap: 6px; }
    .form-group label { font-size: 13px; font-weight: 500; color: #94a3b8; }
    .form-group input {
      padding: 10px 14px; background: #0a1628; border: 1px solid #1e3a5f;
      border-radius: 8px; color: #e2e8f0; font-size: 14px; outline: none;
      transition: border-color 0.15s;
    }
    .form-group input:focus { border-color: #3b82f6; }
    .form-group input::placeholder { color: #475569; }
    .btn-primary {
      padding: 10px 20px; background: #3b82f6; color: white; border: none;
      border-radius: 8px; font-size: 14px; font-weight: 600; cursor: pointer;
      transition: background 0.15s;
    }
    .btn-primary:hover { background: #2563eb; }
    .btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
    .error-message {
      padding: 10px 14px; background: rgba(239, 68, 68, 0.1);
      border: 1px solid rgba(239, 68, 68, 0.3); border-radius: 8px;
      color: #ef4444; font-size: 13px;
    }
    .auth-footer { text-align: center; margin-top: 24px; font-size: 13px; color: #64748b; }
    .auth-footer a { color: #3b82f6; text-decoration: none; }
    .auth-footer a:hover { text-decoration: underline; }
  `]
})
export class RegisterComponent {
  name = '';
  email = '';
  password = '';
  loading = signal(false);
  error = signal('');

  constructor(private authService: AuthService, private router: Router) {}

  onSubmit(): void {
    this.loading.set(true);
    this.error.set('');

    this.authService.register({ name: this.name, email: this.email, password: this.password }).subscribe({
      next: () => {
        this.router.navigate(['/dashboard']);
      },
      error: (err) => {
        this.error.set(err.error?.error || 'Registration failed');
        this.loading.set(false);
      }
    });
  }
}
