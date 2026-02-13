// Share page interactivity: view switching, image preview, selection management

(function () {
    'use strict';

    // State
    let currentView = localStorage.getItem('shareView') || 'grid';
    let selectedFiles = new Set();
    let imageFiles = [];
    let currentImageIndex = 0;

    // DOM elements
    let gridViewBtn, listViewBtn, gridContainer, listContainer;
    let selectAllCheckbox, downloadBtn, fileCountEl;
    let modal, modalImg, modalClose, modalPrev, modalNext;

    // Initialize on page load
    document.addEventListener('DOMContentLoaded', init);

    function init() {
        // Get DOM elements
        gridViewBtn = document.getElementById('gridViewBtn');
        listViewBtn = document.getElementById('listViewBtn');
        gridContainer = document.getElementById('gridView');
        listContainer = document.getElementById('listView');
        selectAllCheckbox = document.getElementById('selectAll');
        downloadBtn = document.getElementById('downloadBtn');
        fileCountEl = document.getElementById('fileCount');
        modal = document.getElementById('imageModal');
        modalImg = document.getElementById('modalImage');
        modalClose = document.getElementById('modalClose');
        modalPrev = document.getElementById('modalPrev');
        modalNext = document.getElementById('modalNext');

        // Set initial view
        setView(currentView);

        // Event listeners
        if (gridViewBtn) gridViewBtn.addEventListener('click', () => setView('grid'));
        if (listViewBtn) listViewBtn.addEventListener('click', () => setView('list'));
        if (selectAllCheckbox) selectAllCheckbox.addEventListener('change', handleSelectAll);
        if (modalClose) modalClose.addEventListener('click', closeModal);
        if (modalPrev) modalPrev.addEventListener('click', showPrevImage);
        if (modalNext) modalNext.addEventListener('click', showNextImage);

        // Setup file checkboxes
        setupCheckboxes();

        // Setup image thumbnails for preview
        setupImagePreviews();

        // Keyboard shortcuts
        document.addEventListener('keydown', handleKeyboard);

        // Update UI
        updateSelectionUI();
    }

    function setView(view) {
        currentView = view;
        localStorage.setItem('shareView', view);

        if (view === 'grid') {
            gridContainer.style.display = 'grid';
            listContainer.style.display = 'none';
            gridViewBtn.classList.add('active');
            listViewBtn.classList.remove('active');
        } else {
            gridContainer.style.display = 'none';
            listContainer.style.display = 'block';
            gridViewBtn.classList.remove('active');
            listViewBtn.classList.add('active');
        }

        // Re-sync checkboxes between views
        syncCheckboxes();
    }

    function setupCheckboxes() {
        const checkboxes = document.querySelectorAll('.fileCheckbox');
        checkboxes.forEach(cb => {
            cb.addEventListener('change', function () {
                const filename = this.value;
                if (this.checked) {
                    selectedFiles.add(filename);
                } else {
                    selectedFiles.delete(filename);
                }
                syncCheckboxes();
                updateSelectionUI();
            });
        });
    }

    function syncCheckboxes() {
        // Sync checkbox states between grid and list views
        const checkboxes = document.querySelectorAll('.fileCheckbox');
        checkboxes.forEach(cb => {
            cb.checked = selectedFiles.has(cb.value);
        });

        // Update select all checkbox
        if (selectAllCheckbox) {
            const allCheckboxes = document.querySelectorAll('.fileCheckbox');
            const allChecked = allCheckboxes.length > 0 &&
                Array.from(allCheckboxes).every(cb => cb.checked);
            selectAllCheckbox.checked = allChecked;
        }
    }

    function handleSelectAll(e) {
        const checkboxes = document.querySelectorAll('.fileCheckbox');
        checkboxes.forEach(cb => {
            cb.checked = e.target.checked;
            if (e.target.checked) {
                selectedFiles.add(cb.value);
            } else {
                selectedFiles.delete(cb.value);
            }
        });
        updateSelectionUI();
    }

    function updateSelectionUI() {
        const count = selectedFiles.size;

        // Update file count
        if (fileCountEl) {
            if (count > 0) {
                fileCountEl.textContent = `${count} selected`;
                fileCountEl.style.display = 'inline';
            } else {
                fileCountEl.style.display = 'none';
            }
        }

        // Update download button
        if (downloadBtn) {
            downloadBtn.disabled = count === 0;
            if (count > 0) {
                downloadBtn.textContent = `📥 Download Selected (${count})`;
            } else {
                downloadBtn.textContent = '📥 Download Selected';
            }
        }
    }

    function setupImagePreviews() {
        // Collect all image elements
        const imgElements = document.querySelectorAll('.preview-thumbnail');

        imgElements.forEach((img, index) => {
            const filename = img.dataset.filename;
            const src = img.src;

            imageFiles.push({ filename, src, element: img });

            // Add click listener to open modal
            img.addEventListener('click', () => openModal(index));
            img.style.cursor = 'pointer';
            img.title = 'Click to preview';
        });
    }

    function openModal(index) {
        if (!modal || !modalImg || imageFiles.length === 0) return;

        currentImageIndex = index;
        const image = imageFiles[index];

        // Load full-size image (replace /thumb/ with /dl/ for higher quality)
        const fullSizeSrc = image.src.replace('/thumb/', '/dl/');
        modalImg.src = fullSizeSrc;
        modalImg.alt = image.filename;

        modal.style.display = 'flex';
        document.body.style.overflow = 'hidden';

        // Show/hide navigation buttons
        updateModalNavigation();
    }

    function closeModal() {
        if (modal) {
            modal.style.display = 'none';
            document.body.style.overflow = '';
        }
    }

    function showPrevImage() {
        if (imageFiles.length === 0) return;
        currentImageIndex = (currentImageIndex - 1 + imageFiles.length) % imageFiles.length;
        const image = imageFiles[currentImageIndex];
        const fullSizeSrc = image.src.replace('/thumb/', '/dl/');
        modalImg.src = fullSizeSrc;
        modalImg.alt = image.filename;
        updateModalNavigation();
    }

    function showNextImage() {
        if (imageFiles.length === 0) return;
        currentImageIndex = (currentImageIndex + 1) % imageFiles.length;
        const image = imageFiles[currentImageIndex];
        const fullSizeSrc = image.src.replace('/thumb/', '/dl/');
        modalImg.src = fullSizeSrc;
        modalImg.alt = image.filename;
        updateModalNavigation();
    }

    function updateModalNavigation() {
        if (modalPrev) modalPrev.style.display = imageFiles.length > 1 ? 'block' : 'none';
        if (modalNext) modalNext.style.display = imageFiles.length > 1 ? 'block' : 'none';
    }

    function handleKeyboard(e) {
        // Modal shortcuts
        if (modal && modal.style.display === 'flex') {
            if (e.key === 'Escape') {
                closeModal();
            } else if (e.key === 'ArrowLeft') {
                showPrevImage();
            } else if (e.key === 'ArrowRight') {
                showNextImage();
            }
        }
    }

    // Click outside modal to close
    if (modal) {
        modal.addEventListener('click', function (e) {
            if (e.target === modal) {
                closeModal();
            }
        });
    }

})();
