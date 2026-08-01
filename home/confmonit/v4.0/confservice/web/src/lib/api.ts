const API_URL = (import.meta.env.VITE_API_URL as string) || 'http://localhost:2020';

export type Parceiro = {
  id: string;
  razaoSocial: string;
  nomeFantasia: string;
  email: string;
  telefone: string;
  software: string;
  webhookUrl: string;
  precoClienteQuinzena: number;
  ativo: boolean;
};

export type ClienteVinculo = {
  id: string;
  idFranqueado: string;
  idCliente: string;
  nomeCliente: string;
  contaExterna: string;
  precoCongelado: number;
  inicioEm: string;
};

function authHeaders(): HeadersInit {
  const token = localStorage.getItem('cs_token');
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
}

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    ...init,
    headers: { ...authHeaders(), ...(init?.headers || {}) },
  });
  const text = await res.text();
  let data: unknown = {};
  try {
    data = text ? JSON.parse(text) : {};
  } catch {
    throw new Error(
      'API nao respondeu JSON (proxy/hospedagem?). Confira se /parceiro/registrar chega na API Go.',
    );
  }
  if (!res.ok) {
    throw new Error((data as { erro?: string }).erro || `HTTP ${res.status}`);
  }
  return data as T;
}

export const api = {
  login: (email: string, senha: string) =>
    req<{ ok: boolean; token: string }>('/parceiro/login', {
      method: 'POST',
      body: JSON.stringify({ email, senha }),
    }),

  registrar: (body: Record<string, unknown>) =>
    req<{ ok: boolean; token: string; parceiro: Parceiro }>('/parceiro/registrar', {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  me: () => req<{ ok: boolean; parceiro: Parceiro }>('/parceiro/me'),

  atualizar: (body: Record<string, unknown>) =>
    req<{ ok: boolean; parceiro: Parceiro }>('/parceiro/me/atualizar', {
      method: 'PUT',
      body: JSON.stringify(body),
    }),

  clientes: () => req<{ ok: boolean; dados: ClienteVinculo[] }>('/parceiro/me/clientes'),

  faturas: () =>
    req<{
      ok: boolean;
      dados: Array<{
        id: string;
        periodoInicio: string;
        periodoFim: string;
        valorBruto: number;
        comissao: number;
        valorLiquido: number;
        status: string;
      }>;
    }>('/parceiro/me/faturas'),

  faturaFechar: (body: Record<string, unknown> = {}) =>
    req<{
      ok: boolean;
      periodoInicio: string;
      periodoFim: string;
      resumo: {
        itens: number;
        faturasFranqueado: number;
        faturasParceiro: number;
        totalBruto: number;
        totalComissao: number;
        totalLiquidoParceiro: number;
      };
    }>('/parceiro/me/fatura/fechar', {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  eventos: (params: { de?: string; ate?: string; status?: string; idCliente?: string; limite?: number } = {}) => {
    const q = new URLSearchParams();
    if (params.de) q.set('de', params.de);
    if (params.ate) q.set('ate', params.ate);
    if (params.status) q.set('status', params.status);
    if (params.idCliente) q.set('idCliente', params.idCliente);
    if (params.limite) q.set('limite', String(params.limite));
    const qs = q.toString();
    return req<{
      ok: boolean;
      resumo: { total: number; ok: number; erro: number; pendente: number; enviando: number };
      dados: Array<{
        id: number;
        createdAt: string;
        updatedAt: string;
        idCliente: string;
        nomeCliente: string;
        idFranqueado: string;
        alarmEventsId?: number;
        status: string;
        tentativas: number;
        httpStatus?: number;
        ultimoErro?: string;
        grupo?: string;
        codigo?: string;
        descricao?: string;
      }>;
    }>(`/parceiro/me/eventos${qs ? `?${qs}` : ''}`);
  },
};
