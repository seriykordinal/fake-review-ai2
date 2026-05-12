export function escapeHtml(str) {
    if (typeof str !== 'string')
        return '';
    return str.replace(/[&<>]/g, (m) => {
        if (m === '&')
            return '&amp;';
        if (m === '<')
            return '&lt;';
        if (m === '>')
            return '&gt;';
        return m;
    });
}
export function formatDate(dateStr) {
    return new Date(dateStr).toLocaleString();
}
export function showMessage(messageDiv, msg, isError = false) {
    messageDiv.textContent = msg;
    messageDiv.style.color = isError ? 'red' : 'green';
    setTimeout(() => (messageDiv.textContent = ''), 4000);
}
//# sourceMappingURL=helpers.js.map