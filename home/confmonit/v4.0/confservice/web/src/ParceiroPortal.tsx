import { useEffect, useState } from 'react';
import type { FormEvent } from 'react';
import { api, type ClienteVinculo, type Parceiro } from './lib/api';
import { formatTelefoneBR, onEmailInput, onTelefoneInput } from './lib/format';
import PasswordField from './components/PasswordField';

type FaturaParceiro = {
  id: string;
  periodoInicio: string;
  periodoFim: string;
  valorBruto: number;
  comissao: number;
  valorLiquido: number;
  status: string;
};

type EventoWebhook = {
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
};

type EventoResumo = {
  total: number;
  ok: number;
  erro: number;
  pendente: number;
  enviando: number;
};
import './ParceiroPortal.css';

type Tab = 'login' | 'register' | 'app';

export default function ParceiroPortal() {
  const [tab, setTab] = useState<Tab>(localStorage.getItem('cs_token') ? 'app' : 'login');
  const [erro, setErro] = useState('');
  const [parceiro, setParceiro] = useState<Parceiro | null>(null);
  const [clientes, setClientes] = useState<ClienteVinculo[]>([]);
  const [faturas, setFaturas] = useState<FaturaParceiro[]>([]);
  const [eventos, setEventos] = useState<EventoWebhook[]>([]);
  const [eventoResumo, setEventoResumo] = useState<EventoResumo | null>(null);
  const [filtroDe, setFiltroDe] = useState('');
  const [filtroAte, setFiltroAte] = useState('');
  const [filtroStatus, setFiltroStatus] = useState('');
  const [saving, setSaving] = useState(false);
  const [fechando, setFechando] = useState(false);
  const [msgOk, setMsgOk] = useState('');

  async function carregarEventos(extra?: { de?: string; ate?: string; status?: string }) {
    const fat = await api.eventos({
      de: extra?.de ?? filtroDe,
      ate: extra?.ate ?? filtroAte,
      status: extra?.status ?? filtroStatus,
      limite: 200,
    });
    setEventos(fat.dados || []);
    setEventoResumo(fat.resumo || null);
  }

  async function carregar() {
    setErro('');
    try {
      const me = await api.me();
      setParceiro(me.parceiro);
      const cli = await api.clientes();
      setClientes(cli.dados || []);
      const fat = await api.faturas();
      setFaturas(fat.dados || []);
      await carregarEventos({ de: '', ate: '', status: '' });
      setTab('app');
    } catch (e) {
      localStorage.removeItem('cs_token');
      setTab('login');
      setErro(e instanceof Error ? e.message : 'sessao expirada');
    }
  }

  useEffect(() => {
    if (localStorage.getItem('cs_token')) void carregar();
  }, []);

  async function onLogin(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setErro('');
    const fd = new FormData(e.currentTarget);
    try {
      const r = await api.login(String(fd.get('email')), String(fd.get('senha')));
      localStorage.setItem('cs_token', r.token);
      await carregar();
    } catch (err) {
      setErro(err instanceof Error ? err.message : 'falha no login');
    }
  }

  async function onRegister(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setErro('');
    const fd = new FormData(e.currentTarget);
    try {
      const r = await api.registrar({
        razaoSocial: fd.get('razaoSocial'),
        nomeFantasia: fd.get('nomeFantasia'),
        email: fd.get('email'),
        telefone: fd.get('telefone'),
        software: fd.get('software'),
        webhookUrl: fd.get('webhookUrl'),
        webhookToken: fd.get('webhookToken'),
        precoClienteQuinzena: Number(fd.get('preco') || 0),
        senha: fd.get('senha'),
      });
      localStorage.setItem('cs_token', r.token);
      await carregar();
    } catch (err) {
      setErro(err instanceof Error ? err.message : 'falha no cadastro');
    }
  }

  async function onSave(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!parceiro) return;
    setSaving(true);
    setErro('');
    const fd = new FormData(e.currentTarget);
    try {
      const r = await api.atualizar({
        nomeFantasia: fd.get('nomeFantasia'),
        telefone: fd.get('telefone'),
        software: fd.get('software'),
        webhookUrl: fd.get('webhookUrl'),
        webhookToken: fd.get('webhookToken'),
        precoClienteQuinzena: Number(fd.get('preco') || 0),
      });
      setParceiro(r.parceiro);
    } catch (err) {
      setErro(err instanceof Error ? err.message : 'falha ao salvar');
    } finally {
      setSaving(false);
    }
  }

  function logout() {
    localStorage.removeItem('cs_token');
    setParceiro(null);
    setClientes([]);
    setEventos([]);
    setEventoResumo(null);
    setTab('login');
  }

  if (tab === 'login') {
    return (
      <div className="shell">
        <header className="hero">
          <p className="brand">ConfService</p>
          <h1>Portal do parceiro</h1>
          <p className="lead">Receba eventos de monitoramento via webhook e gerencie seu preço por cliente.</p>
          <p className="muted">
            <a href="/" style={{ color: 'var(--accent)' }}>
              ← Voltar ao marketplace
            </a>
          </p>
        </header>
        <form className="card" onSubmit={onLogin}>
          <label>
            E-mail
            <input
              name="email"
              type="email"
              required
              autoComplete="email"
              inputMode="email"
              placeholder="usuario@empresa.com"
              onInput={onEmailInput}
            />
          </label>
          <PasswordField name="senha" required minLength={6} autoComplete="current-password" />
          {erro && <p className="erro">{erro}</p>}
          <button type="submit">Entrar</button>
          <button type="button" className="ghost" onClick={() => setTab('register')}>
            Criar conta de parceiro
          </button>
        </form>
      </div>
    );
  }

  if (tab === 'register') {
    return (
      <div className="shell">
        <header className="hero">
          <p className="brand">ConfService</p>
          <h1>Cadastro de parceiro</h1>
        </header>
        <form className="card wide" onSubmit={onRegister}>
          <div className="grid2">
            <label>
              Razão social
              <input name="razaoSocial" required />
            </label>
            <label>
              Nome fantasia
              <input name="nomeFantasia" />
            </label>
            <label>
              E-mail
              <input
                name="email"
                type="email"
                required
                autoComplete="email"
                inputMode="email"
                placeholder="usuario@empresa.com"
                onInput={onEmailInput}
              />
            </label>
            <label>
              Telefone
              <input
                name="telefone"
                type="tel"
                inputMode="numeric"
                autoComplete="tel"
                placeholder="(19) 9 9999-9999"
                maxLength={16}
                onInput={onTelefoneInput}
              />
            </label>
            <label>
              Software
              <select name="software" defaultValue="GENERICO">
                <option value="GENERICO">Genérico</option>
                <option value="MONI">Moni</option>
                <option value="SEGWARE">Segware</option>
                <option value="DGUARD">dGuard</option>
              </select>
            </label>
            <label>
              Preço / cliente / quinzena (R$)
              <input name="preco" type="number" step="0.01" min="0" defaultValue={20} required />
            </label>
            <label className="span2">
              Webhook URL
              <input name="webhookUrl" placeholder="https://seu-sistema.com/webhook/confmonit" />
            </label>
            <label>
              Token do webhook
              <input name="webhookToken" />
            </label>
            <PasswordField name="senha" required minLength={6} autoComplete="new-password" />
          </div>
          {erro && <p className="erro">{erro}</p>}
          <button type="submit">Cadastrar</button>
          <button type="button" className="ghost" onClick={() => setTab('login')}>
            Já tenho conta
          </button>
        </form>
      </div>
    );
  }

  return (
    <div className="shell app">
      <header className="top">
        <div>
          <p className="brand">ConfService</p>
          <h1>{parceiro?.nomeFantasia || parceiro?.razaoSocial}</h1>
        </div>
        <button type="button" className="ghost" onClick={logout}>
          Sair
        </button>
      </header>

      {erro && <p className="erro banner">{erro}</p>}

      <section className="card wide">
        <h2>Dados e webhook</h2>
        {parceiro && (
          <form onSubmit={onSave}>
            <div className="grid2">
              <label>
                Nome fantasia
                <input name="nomeFantasia" defaultValue={parceiro.nomeFantasia} />
              </label>
              <label>
                Telefone
                <input
                  name="telefone"
                  type="tel"
                  inputMode="numeric"
                  autoComplete="tel"
                  placeholder="(19) 9 9999-9999"
                  maxLength={16}
                  defaultValue={formatTelefoneBR(parceiro.telefone || '')}
                  onInput={onTelefoneInput}
                />
              </label>
              <label>
                Software
                <select name="software" defaultValue={parceiro.software}>
                  <option value="GENERICO">Genérico</option>
                  <option value="MONI">Moni</option>
                  <option value="SEGWARE">Segware</option>
                  <option value="DGUARD">dGuard</option>
                </select>
              </label>
              <label>
                Preço / cliente / quinzena (R$)
                <input
                  name="preco"
                  type="number"
                  step="0.01"
                  min="0"
                  defaultValue={parceiro.precoClienteQuinzena}
                  required
                />
              </label>
              <label className="span2">
                Webhook URL
                <input name="webhookUrl" defaultValue={parceiro.webhookUrl} />
              </label>
              <label className="span2">
                Token do webhook
                <input name="webhookToken" placeholder="Bearer / X-Webhook-Token" />
              </label>
            </div>
            <button type="submit" disabled={saving}>
              {saving ? 'Salvando…' : 'Salvar'}
            </button>
          </form>
        )}
      </section>

      <section className="card wide">
        <h2>Clientes vinculados ({clientes.length})</h2>
        {clientes.length === 0 ? (
          <p className="muted">Nenhum cliente ainda. O franqueado escolhe você no Franqueado Pro.</p>
        ) : (
          <table>
            <thead>
              <tr>
                <th>Cliente</th>
                <th>ID</th>
                <th>Franqueado</th>
                <th>Conta externa</th>
                <th>Preço</th>
                <th>Início</th>
              </tr>
            </thead>
            <tbody>
              {clientes.map((c) => (
                <tr key={c.id}>
                  <td>{c.nomeCliente || '—'}</td>
                  <td>
                    <code>{c.idCliente}</code>
                  </td>
                  <td>
                    <code>{c.idFranqueado}</code>
                  </td>
                  <td>{c.contaExterna || '—'}</td>
                  <td>R$ {Number(c.precoCongelado).toFixed(2)}</td>
                  <td>{c.inicioEm}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      <section className="card wide">
        <h2>Eventos enviados (prova de atendimento)</h2>
        <p className="muted">
          Cada evento que o ConfMonit encaminhou ao seu webhook. Status OK = seu endpoint respondeu 2xx.
        </p>
        <div className="grid2" style={{ marginBottom: '0.75rem' }}>
          <label>
            De
            <input type="date" value={filtroDe} onChange={(e) => setFiltroDe(e.target.value)} />
          </label>
          <label>
            Até
            <input type="date" value={filtroAte} onChange={(e) => setFiltroAte(e.target.value)} />
          </label>
          <label>
            Status
            <select value={filtroStatus} onChange={(e) => setFiltroStatus(e.target.value)}>
              <option value="">Todos</option>
              <option value="OK">OK</option>
              <option value="ERRO">ERRO</option>
              <option value="PENDENTE">PENDENTE</option>
              <option value="ENVIANDO">ENVIANDO</option>
            </select>
          </label>
          <label style={{ display: 'flex', alignItems: 'flex-end' }}>
            <button
              type="button"
              onClick={async () => {
                setErro('');
                try {
                  await carregarEventos();
                } catch (e) {
                  setErro(e instanceof Error ? e.message : 'erro ao carregar eventos');
                }
              }}
            >
              Filtrar
            </button>
          </label>
        </div>
        {eventoResumo && (
          <p className="muted">
            Total {eventoResumo.total} · OK {eventoResumo.ok} · Erro {eventoResumo.erro} · Pendente{' '}
            {eventoResumo.pendente} · Enviando {eventoResumo.enviando}
          </p>
        )}
        {eventos.length === 0 ? (
          <p className="muted">Nenhum evento no período.</p>
        ) : (
          <table>
            <thead>
              <tr>
                <th>Data</th>
                <th>Cliente</th>
                <th>Grupo / Cód.</th>
                <th>Alarm ID</th>
                <th>Status</th>
                <th>HTTP</th>
                <th>Tent.</th>
              </tr>
            </thead>
            <tbody>
              {eventos.map((ev) => (
                <tr key={ev.id}>
                  <td>
                    <small>{new Date(ev.createdAt).toLocaleString('pt-BR')}</small>
                  </td>
                  <td>
                    {ev.nomeCliente || '—'}
                    <br />
                    <code>{ev.idCliente}</code>
                  </td>
                  <td>
                    {ev.grupo || '—'}
                    {ev.codigo ? ` / ${ev.codigo}` : ''}
                    {ev.descricao ? (
                      <>
                        <br />
                        <small className="muted">{ev.descricao}</small>
                      </>
                    ) : null}
                  </td>
                  <td>{ev.alarmEventsId ?? '—'}</td>
                  <td>
                    <strong>{ev.status}</strong>
                    {ev.ultimoErro ? (
                      <>
                        <br />
                        <small className="erro">{ev.ultimoErro}</small>
                      </>
                    ) : null}
                  </td>
                  <td>{ev.httpStatus ?? '—'}</td>
                  <td>{ev.tentativas}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      <section className="card wide">
        <h2>Faturas a receber</h2>
        <div className="row-actions" style={{ marginBottom: '0.75rem' }}>
          <button
            type="button"
            disabled={fechando}
            onClick={async () => {
              if (!window.confirm('Fechar a quinzena atual e gerar sua fatura/cobrança?')) return;
              setFechando(true);
              setErro('');
              setMsgOk('');
              try {
                const r = await api.faturaFechar({});
                const resumo = r.resumo;
                setMsgOk(
                  `Quinzena ${r.periodoInicio} a ${r.periodoFim}: ${resumo.itens} item(ns), líquido R$ ${Number(resumo.totalLiquidoParceiro).toFixed(2)}`,
                );
                const fat = await api.faturas();
                setFaturas(fat.dados || []);
              } catch (e) {
                setErro(e instanceof Error ? e.message : 'erro ao fechar fatura');
              } finally {
                setFechando(false);
              }
            }}
          >
            {fechando ? 'Fechando…' : 'Fechar quinzena / gerar cobrança'}
          </button>
        </div>
        {msgOk && <p className="ok">{msgOk}</p>}
        {faturas.length === 0 ? (
          <p className="muted">Nenhuma fatura quinzenal ainda.</p>
        ) : (
          <table>
            <thead>
              <tr>
                <th>Período</th>
                <th>Bruto</th>
                <th>Comissão</th>
                <th>Líquido</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {faturas.map((f) => (
                <tr key={f.id}>
                  <td>
                    {f.periodoInicio} a {f.periodoFim}
                  </td>
                  <td>R$ {Number(f.valorBruto).toFixed(2)}</td>
                  <td>R$ {Number(f.comissao).toFixed(2)}</td>
                  <td>R$ {Number(f.valorLiquido).toFixed(2)}</td>
                  <td>{f.status}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>
    </div>
  );
}
