const API_URL = (import.meta.env.VITE_API_URL as string) || 'http://localhost:2020';

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers || {}),
    },
  });
  const text = await res.text();
  let data: unknown = {};
  try {
    data = text ? JSON.parse(text) : {};
  } catch {
    throw new Error(
      'API nao respondeu JSON (proxy/hospedagem?). Confira se /marketplace chega na API Go.',
    );
  }
  if (!res.ok) {
    throw new Error((data as { erro?: string }).erro || `HTTP ${res.status}`);
  }
  return data as T;
}

export type Categoria = {
  id: string;
  slug: string;
  nome: string;
  icone: string;
  ordem: number;
};

export type Fabricante = {
  id: string;
  nome: string;
  slug: string;
  descricao: string;
  logoUrl: string;
  comissaoDisponivelPct: number;
};

export type Prestador = {
  id: string;
  tipo: string;
  nome: string;
  nomeFantasia: string;
  cidade: string;
  uf: string;
  regiaoAtendimento: string;
  experienciaAnos: number;
  comissaoPlataformaPct: number;
  disponibilidade: string;
  fotoUrl: string;
  bio: string;
  categorias: Categoria[];
  fabricantes: Fabricante[];
  cnpj?: string;
  email?: string;
  telefone?: string;
};

export const marketplace = {
  categorias: () => req<{ ok: boolean; dados: Categoria[] }>('/marketplace/categorias'),
  fabricantes: () => req<{ ok: boolean; dados: Fabricante[] }>('/marketplace/fabricantes'),
  prestadores: (params: { q?: string; cidade?: string; categoria?: string } = {}) => {
    const q = new URLSearchParams();
    if (params.q) q.set('q', params.q);
    if (params.cidade) q.set('cidade', params.cidade);
    if (params.categoria) q.set('categoria', params.categoria);
    const qs = q.toString();
    return req<{ ok: boolean; dados: Prestador[] }>(`/marketplace/prestadores${qs ? `?${qs}` : ''}`);
  },
  prestador: (id: string) =>
    req<{ ok: boolean; prestador: Prestador }>(`/marketplace/prestadores/${id}`),
  registrarPrestador: (body: Record<string, unknown>) =>
    req<{ ok: boolean; prestador: { id: string } }>('/marketplace/prestador/registrar', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  registrarFabricante: (body: Record<string, unknown>) =>
    req<{ ok: boolean; fabricante: { id: string } }>('/marketplace/fabricante/registrar', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  orcamento: (body: Record<string, unknown>) =>
    req<{ ok: boolean; id: string }>('/marketplace/orcamento', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
};
