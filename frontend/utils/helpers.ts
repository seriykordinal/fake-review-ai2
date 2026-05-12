export function escapeHtml(str: string): string {
  if (typeof str !== 'string') return '';
  return str.replace(/[&<>]/g, (m) => {
    if (m === '&') return '&amp;';
    if (m === '<') return '&lt;';
    if (m === '>') return '&gt;';
    return m;
  });
}

export function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleString();
}

export function showMessage(
  messageDiv: HTMLElement,
  msg: string,
  isError = false
): void {
  messageDiv.textContent = msg;
  messageDiv.style.color = isError ? 'red' : 'green';
  setTimeout(() => (messageDiv.textContent = ''), 4000);
}