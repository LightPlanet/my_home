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
