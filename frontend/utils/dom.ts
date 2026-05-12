// Безопасное получение DOM-элемента с проверкой типа
export function getElement<T extends HTMLElement>(id: string): T {
  const el = document.getElementById(id);
  if (!el) throw new Error(`Element with id "${id}" not found`);
  return el as T;
}

export function getElementOrNull<T extends HTMLElement>(id: string): T | null {
  return document.getElementById(id) as T | null;
}