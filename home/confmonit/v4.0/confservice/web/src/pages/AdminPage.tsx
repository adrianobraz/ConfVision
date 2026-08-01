import { useEffect, useState } from 'react';
import type { FormEvent } from 'react';
import PasswordField from '../components/PasswordField';
import './Marketplace.css';
import '../ParceiroPortal.css';

const API_URL = (import.meta.env.VITE_API_URL as string) || 'http://localhost:2020';

type ParceiroAdmin = {
  id: string;
  razaoSocial: string;
  nomeFantasia: string;
  email: string;
  telefone: string;
  software: string;
  precoClienteQuinzena: number;
  ativo: boolean;
  createdAt: string;
};

type Props = { onNavigate: (path: string) => void };

async function adminReq<T>(path: string, init?: RequestInit): Promise<T> {
  const token = localStorage.getItem('cs_admin_token');
  const res = await fetch(`${API_URL}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(init?.headers || {}),
    },
  });
  const text = await res.text();
  let data: unknown = {};
  try {
    data = text ? JSON.parse(text) : {};
  } catch {
    throw new Error('API nao respondeu JSON');
  }
  if (!res.ok) {
    throw new Error((data as { erro?: string }).erro || `HTTP ${res.status}`);
  }
  return data as T;
}

export default function AdminPage({ onNavigate }: Props) {
  const [authed, setAuthed] = useState(!!localStorage.getItem('cs_admin_token'));
  const [lista, setLista] = useState<ParceiroAdmin[]>([]);
  const [erro, setErro] = useState('');
  const [ok, setOk] = useState('');
  const [resetId, setResetId] = useState<string | null>(null);
  const [novaSenha, setNovaSenha] = useState('');

  async function carregar() {
    setErro('');
    try {
      const r = await adminReq<{ ok: boolean; dados: ParceiroAdmin[] }>('/admin/parceiros');
      setLista(r.dados || []);
      setAuthed(true);
    } catch (e) {
      localStorage.removeItem('cs_admin_token');
      setAuthed(false);
      setErro(e instanceof Error ? e.message : 'sessao expirada');
    }
  }

  useEffect(() => {
    if (localStorage.getItem('cs_admin_token')) void carregar();
  }, []);

  async function onLogin(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setErro('');
    const fd = new FormData(e.currentTarget);
    try {
      const r = await adminReq<{ ok: boolean; token: string }>('/admin/login', {
        method: 'POST',
        body: JSON.stringify({
          usuario: fd.get('usuario'),
          senha: fd.get('senha'),
        }),
      });
      localStorage.setItem('cs_admin_token', r.token);
      await carregar();
    } catch (err) {
      setErro(err instanceof Error ? err.message : 'falha no login');
    }
  }

  async function resetSenha(id: string) {
    setErro('');
    setOk('');
    if (novaSenha.length < 6) {
      setErro('Nova senha deve ter no minimo 6 caracteres');
      return;
    }
    try {
      await adminReq(`/admin/parceiros/${id}/reset-senha`, {
        method: 'POST',
        body: JSON.stringify({ senha: novaSenha }),
      });
      setOk('Senha atualizada.');
      setResetId(null);
      setNovaSenha('');
    } catch (err) {
      setErro(err instanceof Error ? err.message : 'falha ao resetar');
    }
  }

  function logout() {
    localStorage.removeItem('cs_admin_token');
    setAuthed(false);
    setLista([]);
  }

  if (!authed) {
    return (
      <div className="shell">
        <header className="hero">
          <p className="brand">ConfService</p>
          <h1>Admin</h1>
          <p className="lead">Gestão de parceiros — reset de senha.</p>
          <p className="muted">
            <button type="button" className="ghost" onClick={() => onNavigate('/')}>
              ← Marketplace
            </button>
          </p>
        </header>
        <form className="card" onSubmit={onLogin}>
          <label>
            Usuário
            <input name="usuario" required autoComplete="username" defaultValue="admin" />
          </label>
          <PasswordField name="senha" required minLength={1} autoComplete="current-password" />
          {erro && <p className="erro">{erro}</p>}
          <button type="submit">Entrar</button>
        </form>
      </div>
    );
  }

  return (
    <div className="shell app">
      <header className="top">
        <div>
          <p className="brand">ConfService</p>
          <h1>Parceiros</h1>
        </div>
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button type="button" className="ghost" onClick={() => onNavigate('/')}>
            Marketplace
          </button>
          <button type="button" className="ghost" onClick={logout}>
            Sair
          </button>
        </div>
      </header>

      {erro && <p className="erro banner">{erro}</p>}
      {ok && <p className="ok">{ok}</p>}

      <section className="card wide">
        <table>
          <thead>
            <tr>
              <th>Razão / Fantasia</th>
              <th>E-mail</th>
              <th>Software</th>
              <th>Ativo</th>
              <th>Senha</th>
            </tr>
          </thead>
          <tbody>
            {lista.map((p) => (
              <tr key={p.id}>
                <td>
                  <strong>{p.razaoSocial}</strong>
                  {p.nomeFantasia ? <div className="muted">{p.nomeFantasia}</div> : null}
                </td>
                <td>{p.email}</td>
                <td>{p.software}</td>
                <td>{p.ativo ? 'sim' : 'não'}</td>
                <td>
                  {resetId === p.id ? (
                    <div className="admin-reset">
                      <PasswordField
                        label=""
                        value={novaSenha}
                        onChange={(e) => setNovaSenha(e.target.value)}
                        placeholder="nova senha"
                        minLength={6}
                        required
                      />
                      <button type="button" onClick={() => void resetSenha(p.id)}>
                        Salvar
                      </button>
                      <button
                        type="button"
                        className="ghost"
                        onClick={() => {
                          setResetId(null);
                          setNovaSenha('');
                        }}
                      >
                        Cancelar
                      </button>
                    </div>
                  ) : (
                    <button type="button" className="ghost" onClick={() => setResetId(p.id)}>
                      Resetar senha
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {!lista.length && <p className="muted">Nenhum parceiro cadastrado.</p>}
      </section>
    </div>
  );
}
