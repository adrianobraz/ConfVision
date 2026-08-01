import { useEffect, useState } from 'react';
import type { FormEvent } from 'react';
import { marketplace, type Categoria, type Fabricante } from '../lib/marketplace';
import './Marketplace.css';

type Props = { onNavigate: (path: string) => void };

export default function CadastroPrestador({ onNavigate }: Props) {
  const [categorias, setCategorias] = useState<Categoria[]>([]);
  const [fabricantes, setFabricantes] = useState<Fabricante[]>([]);
  const [catsSel, setCatsSel] = useState<string[]>([]);
  const [fabsSel, setFabsSel] = useState<string[]>([]);
  const [erro, setErro] = useState('');
  const [ok, setOk] = useState('');

  useEffect(() => {
    void (async () => {
      try {
        const [c, f] = await Promise.all([marketplace.categorias(), marketplace.fabricantes()]);
        setCategorias(c.dados || []);
        setFabricantes(f.dados || []);
      } catch {
        /* vitrine offline */
      }
    })();
  }, []);

  function toggle(list: string[], id: string, set: (v: string[]) => void) {
    set(list.includes(id) ? list.filter((x) => x !== id) : [...list, id]);
  }

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setErro('');
    setOk('');
    const fd = new FormData(e.currentTarget);
    try {
      await marketplace.registrarPrestador({
        tipo: fd.get('tipo'),
        nome: fd.get('nome'),
        nomeFantasia: fd.get('nomeFantasia'),
        cnpj: fd.get('cnpj'),
        email: fd.get('email'),
        telefone: fd.get('telefone'),
        senha: fd.get('senha'),
        cidade: fd.get('cidade'),
        uf: fd.get('uf'),
        regiaoAtendimento: fd.get('regiaoAtendimento'),
        bio: fd.get('bio'),
        experienciaAnos: Number(fd.get('experienciaAnos') || 0),
        comissaoPlataformaPct: Number(fd.get('comissaoPlataformaPct') || 10),
        disponibilidade: fd.get('disponibilidade'),
        categorias: catsSel,
        fabricantes: fabsSel,
      });
      setOk('Cadastro realizado! Seu perfil já aparece na vitrine.');
      e.currentTarget.reset();
      setCatsSel([]);
      setFabsSel([]);
    } catch (err) {
      setErro(err instanceof Error ? err.message : 'Falha no cadastro');
    }
  }

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
        <h1>Cadastro de profissional / empresa</h1>
        <p className="mk-section-lead">
          Publique seus serviços no marketplace. Informe a comissão que será repassada à plataforma em cada serviço fechado.
        </p>

        <form className="mk-form" onSubmit={onSubmit}>
          <label>
            Tipo
            <select name="tipo" defaultValue="PROFISSIONAL">
              <option value="PROFISSIONAL">Profissional</option>
              <option value="EMPRESA">Empresa</option>
            </select>
          </label>
          <label>
            Nome / Razão social
            <input name="nome" required />
          </label>
          <label>
            Nome fantasia
            <input name="nomeFantasia" />
          </label>
          <label>
            CNPJ (empresas)
            <input name="cnpj" />
          </label>
          <label>
            E-mail
            <input name="email" type="email" required />
          </label>
          <label>
            Telefone
            <input name="telefone" />
          </label>
          <label>
            Cidade
            <input name="cidade" required />
          </label>
          <label>
            UF
            <input name="uf" maxLength={2} required placeholder="SP" />
          </label>
          <label className="span2">
            Região de atendimento
            <input name="regiaoAtendimento" placeholder="Campinas, região metropolitana…" />
          </label>
          <label>
            Anos de experiência
            <input name="experienciaAnos" type="number" min={0} defaultValue={0} />
          </label>
          <label>
            Comissão plataforma (%)
            <input
              name="comissaoPlataformaPct"
              type="number"
              step="0.01"
              min={1}
              max={50}
              defaultValue={10}
              required
            />
          </label>
          <label className="span2">
            Disponibilidade
            <input name="disponibilidade" placeholder="Seg–Sex, orçamentos em 24h" />
          </label>
          <label className="span2">
            Sobre você / empresa
            <textarea name="bio" rows={3} />
          </label>
          <label>
            Senha
            <input name="senha" type="password" required minLength={6} />
          </label>

          <fieldset className="span2">
            <legend>Especialidades</legend>
            <div className="mk-checks">
              {categorias.map((c) => (
                <label key={c.id} className="mk-check">
                  <input
                    type="checkbox"
                    checked={catsSel.includes(c.slug)}
                    onChange={() => toggle(catsSel, c.slug, setCatsSel)}
                  />
                  {c.icone} {c.nome}
                </label>
              ))}
            </div>
          </fieldset>

          <fieldset className="span2">
            <legend>Fabricantes / marcas que domina</legend>
            <div className="mk-checks">
              {fabricantes.map((f) => (
                <label key={f.id} className="mk-check">
                  <input
                    type="checkbox"
                    checked={fabsSel.includes(f.slug)}
                    onChange={() => toggle(fabsSel, f.slug, setFabsSel)}
                  />
                  {f.nome} ({f.comissaoDisponivelPct}% canal)
                </label>
              ))}
            </div>
          </fieldset>

          {erro && <p className="mk-erro span2">{erro}</p>}
          {ok && <p className="mk-ok span2">{ok}</p>}
          <button type="submit" className="span2">
            Publicar no marketplace
          </button>
        </form>
      </section>
    </div>
  );
}
