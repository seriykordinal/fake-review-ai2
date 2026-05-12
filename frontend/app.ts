import { AuthService } from './services/auth.js';
import { ProductService } from './services/product.js';
import { api } from './utils/api.js';
import type { Profile, ReviewResult, AnalyzeProductResponse } from './types/index.js';

// ===== Helpers =====
function qs<T extends HTMLElement = HTMLElement>(id: string): T {
    return document.getElementById(id) as T;
}
function escHtml(s: string): string {
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}
function initials(email: string): string {
    return email ? email[0].toUpperCase() : '?';
}

// WB даты приходят как ISO или как русский текст ("14 мая 2024") — оба варианта показываем нормально
function formatDate(s: string): string {
    if (!s) return '';
    const d = new Date(s);
    if (!isNaN(d.getTime())) {
        return d.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' });
    }
    return s; // уже читаемый текст — вернуть как есть
}

// ===== Toast =====
let toastTimer: ReturnType<typeof setTimeout>;
function toast(msg: string, type: 'ok' | 'err' | 'info' = 'info') {
    const el = qs('toast');
    el.textContent = msg;
    el.className = `toast show ${type === 'ok' ? 'toast-ok' : type === 'err' ? 'toast-err' : ''}`;
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => el.classList.remove('show'), 3200);
}

// ===== State =====
let token: string | null = AuthService.getToken();
let profile: Profile | null = null;

// ===== Field clearing =====
function clearInputById(id: string) {
    const el = document.getElementById(id) as HTMLInputElement | HTMLTextAreaElement | null;
    if (el) el.value = '';
}
function clearError(id: string) {
    const el = document.getElementById(id);
    if (el) { el.textContent = ''; el.classList.add('hidden'); }
}

function clearLoginForm() {
    clearInputById('loginEmail');
    clearInputById('loginPassword');
    clearError('loginError');
}

function clearRegisterForm() {
    clearInputById('regEmail');
    clearInputById('regPassword');
    clearInputById('verifyCode');
    clearError('regError');
    qs('verifyPanel').classList.add('hidden');
}

function clearProductArea() {
    clearInputById('productUrl');
    const results = qs('productResults');
    results.classList.add('hidden');
    results.innerHTML = '';
}

function clearPublicArea() {
    clearInputById('reviewText');
    const r = qs('singleResult');
    r.className = 'inline-result hidden';
    r.textContent = '';
}

// Полная очистка при выходе / смене пользователя
function clearAllUserData() {
    clearProductArea();
    clearPublicArea();
    clearLoginForm();
    clearRegisterForm();
}

// ===== Screens =====
type Screen = 'public' | 'auth' | 'history';

function showScreen(s: Screen) {
    qs('publicContent').classList.toggle('hidden', s !== 'public');
    qs('authContent').classList.toggle('hidden', s !== 'auth');
    qs('historyContent').classList.toggle('hidden', s !== 'history');
}

// ===== Header =====
function renderHeader(p: Profile | null) {
    const guestActions  = qs('guestActions');
    const profileTrigger = qs('profileTrigger');

    if (p) {
        guestActions.classList.add('hidden');
        profileTrigger.classList.remove('hidden');
        qs('headerEmail').textContent  = p.email;
        qs('profileAvatar').textContent = initials(p.email);
        qs('dropEmail').textContent    = p.email;
        qs('dropAvatar').textContent   = initials(p.email);
        const roleMap: Record<string, string> = {
            user: 'Пользователь', admin: 'Администратор', super_admin: 'Супер-Админ'
        };
        qs('dropRole').textContent    = roleMap[p.role] ?? p.role;
        qs('dropCreated').textContent = 'Зарегистрирован: ' + formatDate(p.created_at);
        qs('ddAdminBtn').classList.toggle('hidden', p.role !== 'admin' && p.role !== 'super_admin');
    } else {
        guestActions.classList.remove('hidden');
        profileTrigger.classList.add('hidden');
    }
}

// ===== Dropdown =====
const dropdown = qs('profileDropdown');
qs('profileTrigger').addEventListener('click', (e) => {
    e.stopPropagation();
    dropdown.classList.toggle('hidden');
});
document.addEventListener('click', (e) => {
    if (!dropdown.contains(e.target as Node)) dropdown.classList.add('hidden');
});

// ===== Modals =====
function openModal(id: string) { qs(id).classList.remove('hidden'); }
function closeModal(id: string) { qs(id).classList.add('hidden'); }

document.querySelectorAll('[data-close]').forEach(btn => {
    btn.addEventListener('click', () => closeModal((btn as HTMLElement).dataset.close!));
});
document.querySelectorAll('.modal-overlay').forEach(overlay => {
    overlay.addEventListener('click', (e) => {
        if (e.target === overlay) closeModal(overlay.id);
    });
});

// ===== Auth Buttons =====
qs('openLoginBtn').addEventListener('click', () => { clearLoginForm(); openModal('loginModal'); });
qs('openRegisterBtn').addEventListener('click', () => { clearRegisterForm(); openModal('registerModal'); });
qs('hintLoginBtn').addEventListener('click', () => { clearLoginForm(); openModal('loginModal'); });
qs('hintRegBtn').addEventListener('click', () => { clearRegisterForm(); openModal('registerModal'); });

// ===== Login =====
qs('doLoginBtn').addEventListener('click', async () => {
    const email    = qs<HTMLInputElement>('loginEmail').value.trim();
    const password = qs<HTMLInputElement>('loginPassword').value;
    clearError('loginError');
    try {
        const res = await AuthService.login(email, password);
        token = res.token;
        AuthService.saveToken(token);
        closeModal('loginModal');
        clearLoginForm();
        clearProductArea(); // очищаем данные предыдущего пользователя
        profile = await AuthService.getProfile(token);
        renderHeader(profile);
        showScreen('auth');
        toast('Добро пожаловать!', 'ok');
    } catch (e) {
        const errEl = qs('loginError');
        errEl.textContent = (e as Error).message;
        errEl.classList.remove('hidden');
    }
});

// ===== Register =====
qs('doRegisterBtn').addEventListener('click', async () => {
    const email    = qs<HTMLInputElement>('regEmail').value.trim();
    const password = qs<HTMLInputElement>('regPassword').value;
    clearError('regError');
    try {
        const data = await AuthService.register(email, password);
        if (data.token) {
            token = data.token;
            AuthService.saveToken(token);
            closeModal('registerModal');
            clearRegisterForm();
            clearProductArea();
            profile = await AuthService.getProfile(token);
            renderHeader(profile);
            showScreen('auth');
            toast('Регистрация успешна!', 'ok');
        } else {
            qs('verifyPanel').classList.remove('hidden');
            const errEl = qs('regError');
            errEl.textContent = 'Код подтверждения отправлен на почту';
            errEl.classList.remove('hidden');
        }
    } catch (e) {
        const errEl = qs('regError');
        errEl.textContent = (e as Error).message;
        errEl.classList.remove('hidden');
    }
});

// ===== Verify =====
qs('doVerifyBtn').addEventListener('click', async () => {
    const email = qs<HTMLInputElement>('regEmail').value.trim();
    const code  = qs<HTMLInputElement>('verifyCode').value.trim();
    clearError('regError');
    try {
        const res = await AuthService.verify(email, code);
        token = res.token;
        AuthService.saveToken(token);
        closeModal('registerModal');
        clearRegisterForm();
        clearProductArea();
        profile = await AuthService.getProfile(token);
        renderHeader(profile);
        showScreen('auth');
        toast('Верификация успешна!', 'ok');
    } catch (e) {
        const errEl = qs('regError');
        errEl.textContent = (e as Error).message;
        errEl.classList.remove('hidden');
    }
});

// ===== Logout =====
qs('ddLogoutBtn').addEventListener('click', () => {
    AuthService.removeToken();
    token = null;
    profile = null;
    dropdown.classList.add('hidden');
    clearAllUserData(); // очищаем все поля при выходе
    renderHeader(null);
    showScreen('public');
    toast('Вы вышли из системы');
});

// ===== Delete Account =====
qs('ddDeleteBtn').addEventListener('click', () => { dropdown.classList.add('hidden'); openModal('deleteModal'); });

qs('confirmDeleteBtn').addEventListener('click', async () => {
    if (!token) return;
    try {
        await api.delete('/api/account', token);
        closeModal('deleteModal');
        AuthService.removeToken();
        token = null;
        profile = null;
        clearAllUserData();
        renderHeader(null);
        showScreen('public');
        toast('Аккаунт удалён', 'ok');
    } catch (e) {
        closeModal('deleteModal');
        toast((e as Error).message, 'err');
    }
});

// ===== Admin =====
qs('ddAdminBtn').addEventListener('click', () => { window.location.href = '/admin.html'; });

// ===== Single Review Analysis =====
qs('analyzeBtn').addEventListener('click', async () => {
    const text    = qs<HTMLTextAreaElement>('reviewText').value.trim();
    const resultEl = qs('singleResult');
    if (!text) { toast('Введите текст отзыва', 'err'); return; }

    const btn = qs<HTMLButtonElement>('analyzeBtn');
    btn.disabled = true;
    btn.innerHTML = '<span class="loader"></span>';
    resultEl.className = 'inline-result hidden';

    try {
        const data  = await ProductService.analyzeSingleReview(text);
        const pct   = Math.round(data.fake_probability * 100);
        const level = pct > 70 ? 'high' : pct > 40 ? 'medium' : 'low';
        const label = pct > 70 ? 'Высокая вероятность фейка' : pct > 40 ? 'Средняя вероятность фейка' : 'Скорее всего настоящий';
        resultEl.innerHTML = `${label} — <strong>${pct}%</strong>`;
        resultEl.className  = `inline-result fake-${level}`;
    } catch (e) {
        toast((e as Error).message, 'err');
    } finally {
        btn.disabled = false;
        btn.textContent = 'Проверить отзыв';
    }
});

// ===== Product Analysis =====
let isAnalyzing = false;

qs('analyzeProductBtn').addEventListener('click', async () => {
    if (isAnalyzing) { toast('Анализ уже выполняется...'); return; }
    const url = qs<HTMLInputElement>('productUrl').value.trim();
    if (!url)   { toast('Введите ссылку на товар', 'err'); return; }
    if (!token) { toast('Необходима авторизация', 'err');  return; }

    isAnalyzing = true;
    const btn = qs<HTMLButtonElement>('analyzeProductBtn');
    btn.disabled = true;
    btn.innerHTML = '<span class="loader"></span> Анализ...';
    qs('productResults').classList.add('hidden');
    qs('productResults').innerHTML = '';

    try {
        const data = await ProductService.analyzeProduct(url, token);
        renderProductResults(data);
        qs('productResults').classList.remove('hidden');
    } catch (e) {
        toast((e as Error).message, 'err');
    } finally {
        isAnalyzing = false;
        btn.disabled = false;
        btn.textContent = 'Анализировать';
    }
});

function renderProductResults(data: AnalyzeProductResponse) {
    const container = qs('productResults');
    const avg    = data.average_fake_probability ?? 0;
    const avgPct = Math.round(avg * 100);
    const level  = avgPct > 70 ? 'val-high' : avgPct > 40 ? 'val-med' : 'val-low';

    container.innerHTML = `
        <div class="summary-banner">
            <div class="summary-stat">
                <div class="stat-val">${data.total_reviews}</div>
                <div class="stat-label">Отзывов проверено</div>
            </div>
            <div class="summary-stat">
                <div class="stat-val ${level}">${avgPct}%</div>
                <div class="stat-label">Средний показатель фейка</div>
            </div>
        </div>
        <div class="reviews-list" id="reviewsList"></div>
    `;

    if (!data.results?.length) {
        qs('reviewsList').innerHTML = '<div class="history-empty">Нет текстовых отзывов.</div>';
        return;
    }

    data.results.forEach((rev: ReviewResult, i: number) => {
        const pct    = Math.round((rev.fake_probability ?? 0) * 100);
        const lvl    = pct > 70 ? 'high' : pct > 40 ? 'medium' : 'low';
        const rating = rev.rating ? Number(rev.rating) : 0;
        const stars  = rating > 0 ? '★'.repeat(rating) + '☆'.repeat(5 - rating) : '—';
        const date   = rev.date ? formatDate(String(rev.date)) : '';

        const item = document.createElement('div');
        item.className = 'review-item';
        item.innerHTML = `
            <div class="review-header">
                <span class="review-num">#${i + 1}</span>
                <span class="review-rating">${stars}</span>
                ${date ? `<span class="review-date">${escHtml(date)}</span>` : ''}
                <span class="review-fake-badge badge-${lvl}">${pct}% фейк</span>
                <span class="review-chevron">▼</span>
            </div>
            <div class="review-body">
                ${rev.pros ? `<div class="review-field"><strong>Достоинства:</strong> ${escHtml(rev.pros)}</div>` : ''}
                ${rev.cons ? `<div class="review-field"><strong>Недостатки:</strong> ${escHtml(rev.cons)}</div>` : ''}
                ${rev.text
                    ? `<div class="review-field"><strong>Комментарий:</strong> ${escHtml(rev.text)}</div>`
                    : '<div class="review-field" style="color:var(--text-3)">Нет текста</div>'}
            </div>
        `;
        item.querySelector('.review-header')!.addEventListener('click', () => item.classList.toggle('open'));
        qs('reviewsList').appendChild(item);
    });
}

// ===== History =====
qs('ddHistoryBtn').addEventListener('click', async () => {
    dropdown.classList.add('hidden');
    if (!token) return;
    showScreen('history');
    await loadHistory();
});

qs('backFromHistoryBtn').addEventListener('click', () => showScreen('auth'));

async function loadHistory() {
    const list = qs('historyList');
    list.innerHTML = '<div class="history-empty">Загрузка...</div>';
    try {
        const data = await api.get<{
            analysis_id: number; product_url: string;
            product_id: number; total_reviews: number; analyzed_at: string;
        }[]>('/api/history', token!);

        if (!data.length) {
            list.innerHTML = '<div class="history-empty">Вы ещё не анализировали ни одного товара.</div>';
            return;
        }
        list.innerHTML = '';
        data.forEach((item, i) => {
            const el = document.createElement('button');
            el.className = 'history-item';
            el.innerHTML = `
                <span class="history-id">${i + 1}</span>
                <div class="history-info">
                    <div class="history-url">${escHtml(item.product_url)}</div>
                    <div class="history-meta">Товар #${item.product_id} · ${item.total_reviews} отзывов · ${formatDate(item.analyzed_at)}</div>
                </div>
            `;
            el.addEventListener('click', () => {
                qs<HTMLInputElement>('productUrl').value = item.product_url;
                showScreen('auth');
            });
            list.appendChild(el);
        });
    } catch (e) {
        list.innerHTML = `<div class="history-empty">Ошибка: ${escHtml((e as Error).message)}</div>`;
    }
}

// ===== Init =====
async function init() {
    if (token) {
        try {
            profile = await AuthService.getProfile(token);
            renderHeader(profile);
            showScreen('auth');
        } catch {
            AuthService.removeToken();
            token = null;
            renderHeader(null);
            showScreen('public');
        }
    } else {
        renderHeader(null);
        showScreen('public');
    }
}

init();
