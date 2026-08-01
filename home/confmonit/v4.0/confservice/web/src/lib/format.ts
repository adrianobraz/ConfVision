import type { FormEvent } from 'react';

/** Máscara celular BR: (19) 9 9999-9999 */
export function formatTelefoneBR(value: string): string {
  const d = value.replace(/\D/g, '').slice(0, 11);
  if (d.length === 0) return '';
  if (d.length <= 2) return `(${d}`;
  if (d.length <= 3) return `(${d.slice(0, 2)}) ${d.slice(2)}`;
  if (d.length <= 7) return `(${d.slice(0, 2)}) ${d.slice(2, 3)} ${d.slice(3)}`;
  return `(${d.slice(0, 2)}) ${d.slice(2, 3)} ${d.slice(3, 7)}-${d.slice(7)}`;
}

export function onTelefoneInput(e: FormEvent<HTMLInputElement>) {
  const el = e.currentTarget;
  el.value = formatTelefoneBR(el.value);
}

export function onEmailInput(e: FormEvent<HTMLInputElement>) {
  const el = e.currentTarget;
  el.value = el.value.trimStart().toLowerCase();
}
