-- ConfService Marketplace — vitrine de serviços de segurança eletrônica
-- Banco: confservice (já existente). Prefixo cs_

CREATE TABLE IF NOT EXISTS cs_categoria (
  id CHAR(36) NOT NULL PRIMARY KEY,
  slug VARCHAR(80) NOT NULL,
  nome VARCHAR(120) NOT NULL,
  icone VARCHAR(16) NOT NULL DEFAULT '',
  ordem INT NOT NULL DEFAULT 0,
  ativo TINYINT(1) NOT NULL DEFAULT 1,
  UNIQUE KEY uk_cs_categoria_slug (slug)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS cs_fabricante (
  id CHAR(36) NOT NULL PRIMARY KEY,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  nome VARCHAR(120) NOT NULL,
  slug VARCHAR(80) NOT NULL,
  descricao TEXT NULL,
  logo_url VARCHAR(500) NULL,
  comissao_disponivel_pct DECIMAL(5,2) NOT NULL DEFAULT 0 COMMENT '% que o fabricante disponibiliza / repassa',
  email VARCHAR(180) NULL,
  senha_hash VARCHAR(255) NULL,
  ativo TINYINT(1) NOT NULL DEFAULT 1,
  UNIQUE KEY uk_cs_fabricante_slug (slug),
  UNIQUE KEY uk_cs_fabricante_email (email)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS cs_prestador (
  id CHAR(36) NOT NULL PRIMARY KEY,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  tipo ENUM('PROFISSIONAL','EMPRESA') NOT NULL DEFAULT 'PROFISSIONAL',
  nome VARCHAR(180) NOT NULL,
  nome_fantasia VARCHAR(180) NULL,
  cnpj VARCHAR(20) NULL,
  email VARCHAR(180) NOT NULL,
  telefone VARCHAR(30) NULL,
  senha_hash VARCHAR(255) NOT NULL,
  cidade VARCHAR(120) NOT NULL DEFAULT '',
  uf CHAR(2) NOT NULL DEFAULT '',
  regiao_atendimento VARCHAR(255) NULL,
  bio TEXT NULL,
  experiencia_anos INT NOT NULL DEFAULT 0,
  comissao_plataforma_pct DECIMAL(5,2) NOT NULL DEFAULT 10 COMMENT '% repassado à ConfService no serviço fechado',
  disponibilidade VARCHAR(120) NULL,
  foto_url VARCHAR(500) NULL,
  ativo TINYINT(1) NOT NULL DEFAULT 1,
  UNIQUE KEY uk_cs_prestador_email (email),
  KEY idx_cs_prestador_cidade (cidade, uf),
  KEY idx_cs_prestador_ativo (ativo)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS cs_prestador_categoria (
  id_prestador CHAR(36) NOT NULL,
  id_categoria CHAR(36) NOT NULL,
  PRIMARY KEY (id_prestador, id_categoria),
  CONSTRAINT fk_cs_pc_prestador FOREIGN KEY (id_prestador) REFERENCES cs_prestador(id) ON DELETE CASCADE,
  CONSTRAINT fk_cs_pc_categoria FOREIGN KEY (id_categoria) REFERENCES cs_categoria(id) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS cs_prestador_fabricante (
  id_prestador CHAR(36) NOT NULL,
  id_fabricante CHAR(36) NOT NULL,
  PRIMARY KEY (id_prestador, id_fabricante),
  CONSTRAINT fk_cs_pf_prestador FOREIGN KEY (id_prestador) REFERENCES cs_prestador(id) ON DELETE CASCADE,
  CONSTRAINT fk_cs_pf_fabricante FOREIGN KEY (id_fabricante) REFERENCES cs_fabricante(id) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS cs_orcamento_pedido (
  id CHAR(36) NOT NULL PRIMARY KEY,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  texto TEXT NOT NULL,
  cidade VARCHAR(120) NOT NULL DEFAULT '',
  uf CHAR(2) NOT NULL DEFAULT '',
  id_categoria CHAR(36) NULL,
  nome_contato VARCHAR(180) NOT NULL,
  email VARCHAR(180) NOT NULL,
  telefone VARCHAR(30) NULL,
  status ENUM('NOVO','EM_CONTATO','FECHADO','CANCELADO') NOT NULL DEFAULT 'NOVO',
  KEY idx_cs_orc_status (status, created_at),
  CONSTRAINT fk_cs_orc_categoria FOREIGN KEY (id_categoria) REFERENCES cs_categoria(id) ON DELETE SET NULL
) ENGINE=InnoDB;

-- Seed categorias
INSERT IGNORE INTO cs_categoria (id, slug, nome, icone, ordem) VALUES
('c1000000-0001-4000-8000-000000000001', 'instalador-cftv', 'Instalador de CFTV', '📹', 1),
('c1000000-0001-4000-8000-000000000002', 'tecnico-seguranca-eletronica', 'Técnico de segurança eletrônica', '🔐', 2),
('c1000000-0001-4000-8000-000000000003', 'instalacao-alarmes', 'Instalação de alarmes', '🚨', 3),
('c1000000-0001-4000-8000-000000000004', 'monitoramento-24h', 'Monitoramento 24 horas', '📡', 4),
('c1000000-0001-4000-8000-000000000005', 'controle-acesso', 'Controle de acesso', '🚪', 5),
('c1000000-0001-4000-8000-000000000006', 'automacao-portoes', 'Automação e portões', '🏠', 6),
('c1000000-0001-4000-8000-000000000007', 'redes-infraestrutura', 'Redes e infraestrutura', '🛜', 7),
('c1000000-0001-4000-8000-000000000008', 'tecnico-manutencao', 'Técnico / manutenção', '👨‍🔧', 8),
('c1000000-0001-4000-8000-000000000009', 'vigilante', 'Vigilante', '👮', 9),
('c1000000-0001-4000-8000-000000000010', 'seguranca-patrimonial', 'Segurança patrimonial', '🛡️', 10),
('c1000000-0001-4000-8000-000000000011', 'portaria', 'Portaria', '🚪', 11),
('c1000000-0001-4000-8000-000000000012', 'seguranca-armada', 'Segurança armada', '🔫', 12),
('c1000000-0001-4000-8000-000000000013', 'seguranca-eventos', 'Segurança para eventos', '🚗', 13),
('c1000000-0001-4000-8000-000000000014', 'empresas-seguranca', 'Empresas de segurança', '🏢', 14);

-- Seed fabricantes (vitrine; sem login até cadastrarem senha)
INSERT IGNORE INTO cs_fabricante (id, nome, slug, descricao, comissao_disponivel_pct, ativo) VALUES
('f1000000-0001-4000-8000-000000000001', 'Intelbras', 'intelbras', 'Equipamentos de CFTV, alarmes, redes e controle de acesso.', 5.00, 1),
('f1000000-0001-4000-8000-000000000002', 'JFL', 'jfl', 'Alarmes, centrais e soluções de monitoramento.', 5.00, 1),
('f1000000-0001-4000-8000-000000000003', 'Hikvision', 'hikvision', 'Câmeras e soluções de videomonitoramento.', 4.00, 1),
('f1000000-0001-4000-8000-000000000004', 'Outras marcas', 'outras-marcas', 'Outras marcas e equipamentos homologados.', 3.00, 1);
