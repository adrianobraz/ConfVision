import { useEffect, useState } from 'react';
import type { FormEvent } from 'react';
import {
  marketplace,
  type Categoria,
  type Fabricante,
  type Prestador,
} from '../lib/marketplace';
import './Marketplace.css';

type Props = {
  onNavigate: (path: string) => void;
  initialCategoria?: string;
};

export default function MarketplaceHome({ onNavigate, initialCategoria = '' }: Props) {
  const [categorias, setCategorias] = useState<Categoria[]>([]);
  const [fabricantes, setFabricantes] = useState<Fabricante[]>([]);
  const [prestadores, setPrestadores] = useState<Prestador[]>([]);
  const [q, setQ] = useState('');
  const [cidade, setCidade] = useState('');
  const [categoria, setCategoria] = useState(initialCategoria);
  const [loading, setLoading] = useState(true);
  const [erro, setErro] = useState('');
  const [orcOk, setOrcOk] = useState('');
  const [orcErro, setOrcErro] = useState('');

  async function carregar(filtros?: { q?: string; cidade?: string; categoria?: string }) {
    setLoading(true);
    setErro('');
    try {
      const [cats, fabs, prest] = await Promise.all([
        marketplace.categorias(),
        marketplace.fabricantes(),
        marketplace.prestadores({
          q: filtros?.q ?? q,
          cidade: filtros?.cidade ?? cidade,
          categoria: filtros?.categoria ?? categoria,
        }),
      ]);
      setCategorias(cats.dados || []);
      setFabricantes(fabs.dados || []);
      setPrestadores(prest.dados || []);
    } catch (e) {
      setErro(e instanceof Error ? e.message : 'Falha ao carregar marketplace');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void carregar({ categoria: initialCategoria });
  }, [initialCategoria]);

  function onBuscar(e: FormEvent) {
    e.preventDefault();
    void carregar();
  }

  async function onOrcamento(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setOrcOk('');
    setOrcErro('');
    const fd = new FormData(e.currentTarget);
    try {
      await marketplace.orcamento({
        texto: fd.get('texto'),
        cidade: fd.get('cidade'),
        uf: fd.get('uf'),
        categoria: fd.get('categoria') || undefined,
        nomeContato: fd.get('nomeContato'),
        email: fd.get('email'),
        telefone: fd.get('telefone'),
      });
      setOrcOk('Pedido enviado! Em breve profissionais da região entram em contato.');
      e.currentTarget.reset();
    } catch (err) {
      setOrcErro(err instanceof Error ? err.message : 'Falha ao enviar');
    }
  }

  return (
    <div className="mk">
      <header className="mk-nav">
        <button type="button" className="mk-logo" onClick={() => onNavigate('/')}>
          ConfService
        </button>
        <nav>
          <button type="button" className="mk-link" onClick={() => onNavigate('/cadastro')}>
            Sou profissional
          </button>
          <button type="button" className="mk-link" onClick={() => onNavigate('/cadastro-fabricante')}>
            Sou fabricante
          </button>
          <button type="button" className="mk-cta-nav" onClick={() => onNavigate('/parceiro')}>
            Portal monitoramento
          </button>
        </nav>
      </header>

      <section className="mk-hero">
        <div className="mk-hero-bg" aria-hidden />
        <div className="mk-hero-inner">
          <p className="mk-brand">ConfService</p>
          <h1>O marketplace da segurança eletrônica</h1>
          <p className="mk-hero-lead">
            Encontre profissionais e empresas perto de você — CFTV, alarmes, monitoramento e mais.
          </p>
          <form className="mk-search" onSubmit={onBuscar}>
            <input
              name="q"
              value={q}
              onChange={(e) => setQ(e.target.value)}
              placeholder="Ex.: preciso instalar 8 câmeras em um condomínio"
              aria-label="O que você precisa"
            />
            <input
              name="cidade"
              value={cidade}
              onChange={(e) => setCidade(e.target.value)}
              placeholder="Cidade"
              aria-label="Cidade"
            />
            <button type="submit">Buscar</button>
          </form>
        </div>
      </section>

      <section className="mk-section">
        <h2>Categorias</h2>
        <p className="mk-section-lead">Escolha o tipo de serviço que você precisa.</p>
        <div className="mk-cats">
          {categorias.map((c) => (
            <button
              key={c.id}
              type="button"
              className={`mk-cat${categoria === c.slug ? ' active' : ''}`}
              onClick={() => {
                const next = categoria === c.slug ? '' : c.slug;
                setCategoria(next);
                void carregar({ categoria: next });
              }}
            >
              <span className="mk-cat-ico">{c.icone}</span>
              <span>{c.nome}</span>
            </button>
          ))}
        </div>
      </section>

      <section className="mk-section mk-section-alt">
        <h2>Profissionais e empresas</h2>
        <p className="mk-section-lead">
          {loading
            ? 'Carregando…'
            : prestadores.length
              ? `${prestadores.length} resultado(s)`
              : 'Nenhum cadastro ainda — seja o primeiro da sua região.'}
        </p>
        {erro && <p className="mk-erro">{erro}</p>}
        <div className="mk-grid">
          {prestadores.map((p) => (
            <article key={p.id} className="mk-card">
              <div className="mk-card-top">
                <span className="mk-badge">{p.tipo === 'EMPRESA' ? 'Empresa' : 'Profissional'}</span>
                <span className="mk-comissao">{p.comissaoPlataformaPct}% plataforma</span>
              </div>
              <h3>{p.nomeFantasia || p.nome}</h3>
              <p className="mk-meta">
                {[p.cidade, p.uf].filter(Boolean).join(' / ') || 'Brasil'}
                {p.experienciaAnos > 0 ? ` · ${p.experienciaAnos} anos` : ''}
              </p>
              {p.bio && <p className="mk-bio">{p.bio}</p>}
              <div className="mk-tags">
                {(p.categorias || []).slice(0, 3).map((c) => (
                  <span key={c.id}>{c.icone} {c.nome}</span>
                ))}
              </div>
              <div className="mk-tags brands">
                {(p.fabricantes || []).map((f) => (
                  <span key={f.id}>{f.nome}</span>
                ))}
              </div>
              <button type="button" className="mk-btn-outline" onClick={() => onNavigate(`/prestador/${p.id}`)}>
                Ver perfil
              </button>
            </article>
          ))}
        </div>
        {!loading && !prestadores.length && (
          <div className="mk-empty">
            <button type="button" onClick={() => onNavigate('/cadastro')}>
              Cadastrar meu serviço
            </button>
          </div>
        )}
      </section>

      <section className="mk-section">
        <h2>Fabricantes</h2>
        <p className="mk-section-lead">Marcas com comissão disponível para o canal.</p>
        <div className="mk-brands">
          {fabricantes.map((f) => (
            <div key={f.id} className="mk-brand-item">
              <strong>{f.nome}</strong>
              <span>Comissão {f.comissaoDisponivelPct}%</span>
              {f.descricao && <p>{f.descricao}</p>}
            </div>
          ))}
        </div>
        <button type="button" className="mk-btn-ghost" onClick={() => onNavigate('/cadastro-fabricante')}>
          Cadastrar minha marca
        </button>
      </section>

      <section className="mk-section mk-orc" id="orcamento">
        <h2>Peça um orçamento</h2>
        <p className="mk-section-lead">Descreva o que precisa — conectamos você a profissionais da região.</p>
        <form className="mk-orc-form" onSubmit={onOrcamento}>
          <label className="span2">
            O que você precisa?
            <textarea name="texto" required rows={3} placeholder="Preciso instalar 8 câmeras em um condomínio em Campinas." />
          </label>
          <label>
            Cidade
            <input name="cidade" />
          </label>
          <label>
            UF
            <input name="uf" maxLength={2} placeholder="SP" />
          </label>
          <label>
            Categoria
            <select name="categoria" defaultValue="">
              <option value="">Qualquer</option>
              {categorias.map((c) => (
                <option key={c.id} value={c.slug}>
                  {c.nome}
                </option>
              ))}
            </select>
          </label>
          <label>
            Seu nome
            <input name="nomeContato" required />
          </label>
          <label>
            E-mail
            <input name="email" type="email" required />
          </label>
          <label>
            Telefone
            <input name="telefone" />
          </label>
          {orcOk && <p className="mk-ok span2">{orcOk}</p>}
          {orcErro && <p className="mk-erro span2">{orcErro}</p>}
          <button type="submit" className="span2">
            Solicitar orçamento
          </button>
        </form>
      </section>

      <footer className="mk-foot">
        <strong>ConfService</strong>
        <span>Marketplace de segurança eletrônica · Monitoramento parceiro</span>
      </footer>
    </div>
  );
}
