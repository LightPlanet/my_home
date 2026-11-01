// ---------- Сворачивание сайдбара ----------
const sidebar = document.getElementById('sidebar');
const toggleBtn = document.getElementById('toggleBtn');

toggleBtn.addEventListener('click', () => {
  sidebar.classList.toggle('collapsed');
  // Запоминаем состояние
  localStorage.setItem('sidebarCollapsed', sidebar.classList.contains('collapsed'));
});
if (localStorage.getItem('sidebarCollapsed') === 'true') {
  sidebar.classList.add('collapsed');
}

// ---------- Переключение разделов ----------
const navItems = document.querySelectorAll('.nav-item');
const sections = {
  products: document.getElementById('section-products'),
  recipes:  document.getElementById('section-recipes'),
};
const pageTitle = document.getElementById('pageTitle');
const addBtn = document.getElementById('addBtn');

navItems.forEach(item => {
  item.addEventListener('click', (e) => {
    e.preventDefault();
    navItems.forEach(i => i.classList.remove('active'));
    item.classList.add('active');

    const target = item.dataset.section;
    Object.values(sections).forEach(s => s.classList.add('hidden'));
    if (sections[target]) sections[target].classList.remove('hidden');

    pageTitle.textContent = item.querySelector('.nav-label').textContent;
  });
});

// ---------- Фильтры (пример) ----------
document.querySelectorAll('.filter-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
    btn.classList.add('active');
    // TODO: реализовать фильтрацию строк или запрос к API
  });
});

// ---------- Кнопка "Добавить" ----------
addBtn.addEventListener('click', () => {
  // TODO: открыть модалку / перейти на форму
  alert('Открыть форму добавления (здесь будет модалка)');
});

// ---------- Загрузка продуктов с бэкенда (заготовка) ----------
async function loadProducts() {
  try {
    const res = await fetch('/api/products');
    if (!res.ok) throw new Error('Ошибка загрузки');
    const products = await res.json();
    renderProducts(products);
  } catch (err) {
    console.error(err);
  }
}

function renderProducts(products) {
  const tbody = document.getElementById('productsBody');
  tbody.innerHTML = '';
  const today = new Date();
  products.forEach(p => {
    const exp = new Date(p.expires_at);
    const diffDays = Math.ceil((exp - today) / (1000 * 60 * 60 * 24));
    let badgeClass = 'fresh', badgeText = 'Свежее';
    if (diffDays < 0) { badgeClass = 'expired'; badgeText = 'Просрочен'; }
    else if (diffDays <= 3) { badgeClass = 'expiring'; badgeText = 'Скоро истечёт'; }

    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td>${p.name}</td>
      <td>${p.quantity ?? ''}</td>
      <td>${p.expires_at?.slice(0,10) ?? ''}</td>
      <td><span class="badge ${badgeClass}">${badgeText}</span></td>
      <td class="actions">
        <button class="icon-btn" data-id="${p.id}" data-action="edit">✏️</button>
        <button class="icon-btn" data-id="${p.id}" data-action="delete">🗑️</button>
      </td>`;
    tbody.appendChild(tr);
  });
}

// Раскомментируй, когда будет готов бэкенд:
// loadProducts();
