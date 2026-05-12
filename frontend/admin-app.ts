import { getElement } from './utils/dom.js';
import { escapeHtml, formatDate, showMessage } from './utils/helpers.js';
import { AuthService } from './services/auth.js';
import { AdminService } from './services/admin.js';
import type { User, ProductAnalysis, Stats } from './types/index.js';

const statsDiv         = getElement('stats');
const usersTableBody   = document.querySelector('#usersTable tbody') as HTMLElement;
const productsTableBody = document.querySelector('#productsTable tbody') as HTMLElement;
const refreshUsersBtn  = getElement<HTMLButtonElement>('refreshUsers');
const refreshProductsBtn = getElement<HTMLButtonElement>('refreshProducts');
const backToSiteBtn    = getElement<HTMLButtonElement>('backToSite');
const logoutAdminBtn   = getElement<HTMLButtonElement>('logoutBtn');

let token = AuthService.getToken();
if (!token) window.location.href = '/';

let currentUserRole: User['role'] = 'user';

function showError(msg: string): void {
  alert(msg);
  window.location.href = '/';
}

async function loadStats(): Promise<void> {
  try {
    const stats = await AdminService.getStats(token!);
    statsDiv.innerHTML = `
      <p>
        Пользователей: <strong>${stats.total_users}</strong> &nbsp;|&nbsp;
        Анализов: <strong>${stats.total_analyses}</strong> &nbsp;|&nbsp;
        Отзывов: <strong>${stats.total_reviews_analyzed}</strong> &nbsp;|&nbsp;
        Средний фейк: <strong>${(stats.global_avg_fake * 100).toFixed(1)}%</strong>
      </p>
    `;
  } catch (err) {
    console.error('loadStats:', err);
  }
}

async function loadUsers(): Promise<void> {
  try {
    const users = await AdminService.getUsers(token!);
    usersTableBody.innerHTML = '';
    if (!users.length) {
      usersTableBody.innerHTML = '<tr><td colspan="4">Нет пользователей</td></tr>';
      return;
    }
    for (const user of users) {
      const row = document.createElement('tr');
      row.innerHTML = `
        <td>${user.id}</td>
        <td>${escapeHtml(user.email)}</td>
        <td>${user.role}</td>
        <td>${renderUserActions(user)}</td>
      `;
      usersTableBody.appendChild(row);
    }
    attachUserEvents();
  } catch (err) {
    console.error('loadUsers:', err);
    usersTableBody.innerHTML = '<tr><td colspan="4">Ошибка загрузки пользователей</td></tr>';
  }
}

function renderUserActions(user: User): string {
  if (currentUserRole === 'super_admin') {
    if (user.role === 'super_admin') return 'Супер-админ (нельзя изменять)';
    return `
      <button class="danger" data-id="${user.id}" data-action="deleteUser">Удалить</button>
      <select data-id="${user.id}" class="roleSelect">
        <option value="user"  ${user.role === 'user'  ? 'selected' : ''}>user</option>
        <option value="admin" ${user.role === 'admin' ? 'selected' : ''}>admin</option>
      </select>
      <button class="success" data-id="${user.id}" data-action="changeRole">Сменить роль</button>
    `;
  }
  if (currentUserRole === 'admin') {
    if (user.role !== 'admin' && user.role !== 'super_admin') {
      return `<button class="danger" data-id="${user.id}" data-action="deleteUser">Удалить</button>`;
    }
    return 'Недоступно';
  }
  return '';
}

function attachUserEvents(): void {
  document.querySelectorAll('[data-action="deleteUser"]').forEach(btn => {
    btn.addEventListener('click', async () => {
      const id = (btn as HTMLElement).dataset.id;
      if (!id || !confirm('Удалить пользователя? Это также удалит все его анализы.')) return;
      try {
        await AdminService.deleteUser(Number(id), token!);
        await Promise.all([loadUsers(), loadProducts()]);
      } catch {
        alert('Не удалось удалить пользователя');
      }
    });
  });

  document.querySelectorAll('[data-action="changeRole"]').forEach(btn => {
    btn.addEventListener('click', async () => {
      const id = (btn as HTMLElement).dataset.id;
      if (!id) return;
      const select = document.querySelector(`.roleSelect[data-id="${id}"]`) as HTMLSelectElement;
      const email  = (btn.closest('tr') as HTMLTableRowElement)?.cells[1]?.innerText;
      if (!email) return;
      try {
        await AdminService.changeRole(email, select.value, token!);
        await loadUsers();
      } catch {
        alert('Не удалось сменить роль');
      }
    });
  });
}

async function loadProducts(): Promise<void> {
  try {
    const products = await AdminService.getProducts(token!);
    productsTableBody.innerHTML = '';
    if (!products.length) {
      productsTableBody.innerHTML = '<tr><td colspan="6">Нет анализов товаров</td></tr>';
      return;
    }
    for (const p of products) {
      const row = document.createElement('tr');
      row.innerHTML = `
        <td>${p.analysis_id}</td>
        <td>${escapeHtml(p.user_email)}</td>
        <td>${p.product_id}</td>
        <td>${p.total_reviews}</td>
        <td>${formatDate(p.analyzed_at)}</td>
        <td><button class="danger" data-id="${p.analysis_id}" data-action="deleteProduct">Удалить</button></td>
      `;
      productsTableBody.appendChild(row);
    }
    attachProductEvents();
  } catch (err) {
    console.error('loadProducts:', err);
    productsTableBody.innerHTML = '<tr><td colspan="6">Ошибка загрузки анализов</td></tr>';
  }
}

function attachProductEvents(): void {
  document.querySelectorAll('[data-action="deleteProduct"]').forEach(btn => {
    btn.addEventListener('click', async () => {
      const id = (btn as HTMLElement).dataset.id;
      if (!id || !confirm('Удалить анализ товара? Все отзывы также будут удалены.')) return;
      try {
        await AdminService.deleteProduct(Number(id), token!);
        await loadProducts();
      } catch {
        alert('Не удалось удалить анализ');
      }
    });
  });
}

async function initAdmin(): Promise<void> {
  try {
    const me = await AuthService.getMe(token!);
    currentUserRole = me.role;
    if (currentUserRole !== 'admin' && currentUserRole !== 'super_admin') {
      throw new Error('Forbidden');
    }
    await Promise.all([loadStats(), loadUsers(), loadProducts()]);
  } catch {
    showError('Ошибка авторизации');
    return;
  }

  refreshUsersBtn.onclick  = loadUsers;
  refreshProductsBtn.onclick = loadProducts;
  backToSiteBtn.onclick    = () => { window.location.href = '/'; };
  logoutAdminBtn.onclick   = () => { AuthService.removeToken(); window.location.href = '/'; };
}

initAdmin();
