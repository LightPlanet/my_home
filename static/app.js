// Управление модальным окном рецептов (в стиле вашего существующего окна)
function toggleRecipeModal(show) {
    const modal = document.getElementById('recipeModal');
    if (modal) {
        if (show) {
            modal.classList.add('active');
        } else {
            modal.classList.remove('active');
        }
    }
}

// Динамическое добавление строк для ингредиентов
function addIngredientFormRow() {
    const container = document.getElementById('ingredientsFormContainer');
    if (!container) return;
    
    const row = document.createElement('div');
    row.className = 'ingredient-row';
    row.style.display = 'flex';
    row.style.gap = '8px';
    row.style.alignItems = 'center';
    
    // Инпуты получают идентичные вашим стили оформления
    row.innerHTML = `
        <input type="text" name="ingr_names[]" placeholder="Название" required style="flex-grow: 1; padding: 10px; border: 1px solid var(--border-color); border-radius: 6px;">
        <input type="number" name="ingr_counts[]" value="1" min="1" required style="width: 80px; padding: 10px; border: 1px solid var(--border-color); border-radius: 6px;">
        <button type="button" onclick="this.parentElement.remove()" style="background: none; border: none; color: var(--danger); font-size: 20px; cursor: pointer; padding: 0 4px;">&times;</button>
    `;
    container.appendChild(row);
}

document.addEventListener('DOMContentLoaded', () => {
    console.log("JS скрипт успешно загружен!");

    // --- Логика Сайдбара ---
    const sidebar = document.getElementById('sidebar');
    const toggleSidebar = document.getElementById('toggleSidebar');
    
    // Проверяем ОБА элемента сайдбара перед тем, как вешать события
    if (sidebar && toggleSidebar) {
        if (localStorage.getItem('sidebar-collapsed') === 'true') {
            sidebar.classList.add('collapsed');
        }
        toggleSidebar.addEventListener('click', () => {
            sidebar.classList.toggle('collapsed');
            localStorage.setItem('sidebar-collapsed', sidebar.classList.contains('collapsed'));
        });
        console.log("Сайдбар успешно инициализирован");
    } else {
        console.warn("Элементы сайдбара не найдены на этой странице. Пропускаем...");
    }

    // --- Логика Модального окна ---
    const modal = document.getElementById('productModal');
    const openModalBtn = document.getElementById('openModalBtn');
    const closeModalBtn = document.getElementById('closeModalBtn');
    const cancelModalBtn = document.getElementById('cancelModalBtn');

    // Проверяем элементы модального окна отдельно
    if (modal && openModalBtn) {
        const showModal = () => modal.classList.add('active');
        const hideModal = () => {
            modal.classList.remove('active');
            const form = modal.querySelector('form');
            if (form) form.reset();
        };

        openModalBtn.addEventListener('click', showModal);
        
        if (closeModalBtn) closeModalBtn.addEventListener('click', hideModal);
        if (cancelModalBtn) cancelModalBtn.addEventListener('click', hideModal);

        modal.addEventListener('click', (e) => {
            if (e.target === modal) hideModal();
        });
        
        console.log("Модальное окно успешно инициализировано!");
    } else {
        console.warn("Кнопка добавления или модальное окно не найдены на этой странице. Пропускаем...");
    }
});
