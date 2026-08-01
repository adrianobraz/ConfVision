import { useEffect, useState } from 'react';
import MarketplaceHome from './pages/MarketplaceHome';
import CadastroPrestador from './pages/CadastroPrestador';
import CadastroFabricante from './pages/CadastroFabricante';
import PrestadorDetalhe from './pages/PrestadorDetalhe';
import ParceiroPortal from './ParceiroPortal';
import AdminPage from './pages/AdminPage';

function pathNow() {
  return window.location.pathname || '/';
}

export default function App() {
  const [path, setPath] = useState(pathNow);

  useEffect(() => {
    const onPop = () => setPath(pathNow());
    window.addEventListener('popstate', onPop);
    return () => window.removeEventListener('popstate', onPop);
  }, []);

  function navigate(to: string) {
    const [pathname, hash] = to.split('#');
    const next = pathname || '/';
    if (next !== window.location.pathname) {
      window.history.pushState({}, '', next + (hash ? `#${hash}` : ''));
      setPath(next);
    } else if (hash) {
      window.location.hash = hash;
      setPath(next);
    } else {
      setPath(next);
    }
    if (hash) {
      requestAnimationFrame(() => {
        document.getElementById(hash)?.scrollIntoView({ behavior: 'smooth' });
      });
    } else {
      window.scrollTo(0, 0);
    }
  }

  if (path === '/admin' || path === '/admin/') {
    return <AdminPage onNavigate={navigate} />;
  }

  if (path === '/parceiro' || path.startsWith('/parceiro/')) {
    return <ParceiroPortal />;
  }

  if (path === '/cadastro') {
    return <CadastroPrestador onNavigate={navigate} />;
  }

  if (path === '/cadastro-fabricante') {
    return <CadastroFabricante onNavigate={navigate} />;
  }

  const m = path.match(/^\/prestador\/([^/]+)/);
  if (m) {
    return <PrestadorDetalhe id={m[1]} onNavigate={navigate} />;
  }

  return <MarketplaceHome onNavigate={navigate} />;
}
