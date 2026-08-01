import { useEffect, useState } from 'react';
import { marketplace, type Prestador } from '../lib/marketplace';
import './Marketplace.css';

type Props = { id: string; onNavigate: (path: string) => void };

export default function PrestadorDetalhe({ id, onNavigate }: Props) {
  const [p, setP] = useState<Prestador | null>(null);
  const [erro, setErro] = useState('');

  useEffect(() => {
    void (async () => {
      try {
        const r = await marketplace.prestador(id);
        setP(r.prestador);
      } catch (e) {
        setErro(e instanceof Error ? e.message : 'Não encontrado');
      }
    })();
  }, [id]);

  return (
    <div className="mk mk-form-page">
      <header className="mk-nav">
        <button type="button" className="mk-logo" onClick={() => onNavigate('/')}>
          ConfService
        </button>
        <nav>
          <button type="button" className="mk-link" onClick={() => onNavigate('/')}>
            Voltar à loja
          </button>
        </nav>
      </header>

      <section className="mk-section">
        {erro && <p className="mk-erro">{erro}</p>}
        {p && (
          <>
            <span className="mk-badge">{p.tipo === 'EMPRESA' ? 'Empresa' : 'Profissional'}</span>
            <h1>{p.nomeFantasia || p.nome}</h1>
            <p className="mk-meta">
              {[p.cidade, p.uf].filter(Boolean).join(' / ')}
              {p.experienciaAnos > 0 ? ` · ${p.experienciaAnos} anos de experiência` : ''}
            </p>
            {p.bio && <p className="mk-bio-lg">{p.bio}</p>}
            <dl className="mk-dl">
              <div>
                <dt>Comissão plataforma</dt>
                <dd>{p.comissaoPlataformaPct}%</dd>
              </div>
              {p.disponibilidade && (
                <div>
                  <dt>Disponibilidade</dt>
                  <dd>{p.disponibilidade}</dd>
                </div>
              )}
              {p.regiaoAtendimento && (
                <div>
                  <dt>Área de atendimento</dt>
                  <dd>{p.regiaoAtendimento}</dd>
                </div>
              )}
              {p.telefone && (
                <div>
                  <dt>Telefone</dt>
                  <dd>{p.telefone}</dd>
                </div>
              )}
            </dl>
            <h2>Especialidades</h2>
            <div className="mk-tags">
              {(p.categorias || []).map((c) => (
                <span key={c.id}>
                  {c.icone} {c.nome}
                </span>
              ))}
            </div>
            <h2>Fabricantes</h2>
            <div className="mk-tags brands">
              {(p.fabricantes || []).map((f) => (
                <span key={f.id}>
                  {f.nome} · canal {f.comissaoDisponivelPct}%
                </span>
              ))}
            </div>
            <button type="button" onClick={() => onNavigate('/#orcamento')}>
              Solicitar orçamento
            </button>
          </>
        )}
      </section>
    </div>
  );
}
