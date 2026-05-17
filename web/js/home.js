const revealSections = document.querySelectorAll('.reveal-section');

if ('IntersectionObserver' in window) {
  const observer = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          entry.target.classList.add('is-visible');
          observer.unobserve(entry.target);
        }
      });
    },
    { threshold: 0.16 },
  );

  revealSections.forEach((section) => observer.observe(section));
} else {
  revealSections.forEach((section) => section.classList.add('is-visible'));
}
