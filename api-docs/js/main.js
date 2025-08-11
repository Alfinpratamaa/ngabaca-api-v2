document.addEventListener("DOMContentLoaded", () => {
  const contentEl = document.getElementById("content");
  const navLinks = document.querySelectorAll(".nav-link");

  const loadContent = async (page) => {
    // Tampilkan loading
    contentEl.innerHTML =
      '<p class="text-center mt-20 text-gray-500">Loading...</p>';

    try {
      const response = await fetch(`./content/${page}.html`);
      if (!response.ok) throw new Error("Page not found");

      const html = await response.text();
      contentEl.innerHTML = html;

      // Setelah konten baru dimuat, aktifkan kembali tombol copy
      initializeCopyButtons();
    } catch (error) {
      contentEl.innerHTML = `<p class="text-center mt-20 text-red-500">Error: Could not load content for '${page}'.</p>`;
    }
  };

  const initializeCopyButtons = () => {
    const copyButtons = contentEl.querySelectorAll(".copy-btn");
    copyButtons.forEach((button) => {
      button.addEventListener("click", () => {
        const codeBlock = button.previousElementSibling.querySelector("code");
        navigator.clipboard.writeText(codeBlock.innerText).then(() => {
          button.innerText = "Copied!";
          setTimeout(() => {
            button.innerText = "Copy";
          }, 2000);
        });
      });
    });
  };

  const handleRouteChange = () => {
    // Ambil hash dari URL, atau default ke 'auth'
    const page = window.location.hash.substring(1) || "auth";

    // Tandai link yang aktif
    navLinks.forEach((link) => {
      link.classList.toggle("active", link.getAttribute("href") === `#${page}`);
    });

    loadContent(page);
  };

  // Dengarkan perubahan hash (saat link di-klik)
  window.addEventListener("hashchange", handleRouteChange);

  // Muat konten awal saat halaman pertama kali dibuka
  handleRouteChange();
});
