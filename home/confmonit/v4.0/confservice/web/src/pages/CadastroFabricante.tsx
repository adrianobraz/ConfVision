import { useState } from 'react';
import type { FormEvent } from 'react';
import { marketplace } from '../lib/marketplace';
import './Marketplace.css';

type Props = { onNavigate: (path: string) => void };

export default function CadastroFabricante({ onNavigate }: Props) {
  const [erro, setErro] = useState('');
  const [ok, setOk] = useState('');

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setErro('');
    setOk('');
    const fd = new FormData(e.currentTarget);
    try {
      await marketplace.registrarFabricante({
        nome: fd.get('nome'),
        slug: fd.get('slug'),
        descricao: fd.get('descricao'),
        logoUrl: fd.get('logoUrl'),
        comissaoDisponivelPct: Number(fd.get('comissaoDisponivelPct') || 0),
        email: fd.get('email'),
        senha: fd.get('senha'),
      });
      setOk('Fabricante cadastrado! Sua marca e comissão aparecem na vitrine.');
      e.currentTarget.reset();
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
        <h1>Cadastro de fabricante</h1>
        <p className="mk-section-lead">
          Cadastre sua marca (Intelbras, JFL, etc.) e informe a comissão disponível para o canal ConfService.
        </p>

        <form className="mk-form" onSubmit={onSubmit}>
          <label>
            Nome da marca
            <input name="nome" required placeholder="Intelbras" />
          </label>
          <label>
            Slug (opcional)
            <input name="slug" placeholder="intelbras" />
          </label>
          <label className="span2">
            Descrição
            <textarea name="descricao" rows={3} />
          </label>
          <label>
            Logo URL
            <input name="logoUrl" placeholder="https://…" />
          </label>
          <label>
            Comissão disponível (%)
            <input
              name="comissaoDisponivelPct"
              type="number"
              step="0.01"
              min={0}
              max={50}
              defaultValue={5}
              required
            />
          </label>
          <label>
            E-mail
            <input name="email" type="email" required />
          </label>
          <label>
            Senha
            <input name="senha" type="password" required minLength={6} />
          </label>
          {erro && <p className="mk-erro span2">{erro}</p>}
          {ok && <p className="mk-ok span2">{ok}</p>}
          <button type="submit" className="span2">
            Cadastrar fabricante
          </button>
        </form>
      </section>
    </div>
  );
}
