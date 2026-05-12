// Безопасное получение DOM-элемента с проверкой типа
export function getElement(id) {
    const el = document.getElementById(id);
    if (!el)
        throw new Error(`Element with id "${id}" not found`);
    return el;
}
export function getElementOrNull(id) {
    return document.getElementById(id);
}
//# sourceMappingURL=dom.js.map