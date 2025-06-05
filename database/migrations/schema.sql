-- Script di inizializzazione del database
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- Creazione del database
CREATE DATABASE IF NOT EXISTS `ristorante` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
USE `ristorante`;

-- Tabella Ristorante (informazioni generali del ristorante)
CREATE TABLE IF NOT EXISTS `ristorante` (
  `id_ristorante` INT NOT NULL AUTO_INCREMENT,
  `nome` VARCHAR(100) NOT NULL,
  `numero_tavoli` INT NOT NULL,
  `costo_coperto` DECIMAL(10,2) NOT NULL,
  PRIMARY KEY (`id_ristorante`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Tabella Tavolo 
CREATE TABLE IF NOT EXISTS `tavolo` (
  `id_tavolo` INT NOT NULL AUTO_INCREMENT,
  `max_posti` INT NOT NULL,
  `stato` ENUM('libero', 'occupato') NOT NULL DEFAULT 'libero',
  `id_ristorante` INT NOT NULL,
  PRIMARY KEY (`id_tavolo`),
  FOREIGN KEY (`id_ristorante`) REFERENCES `ristorante` (`id_ristorante`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Tabella Ingrediente
CREATE TABLE IF NOT EXISTS `ingrediente` (
  `id_ingrediente` INT NOT NULL AUTO_INCREMENT,
  `nome` VARCHAR(100) NOT NULL,
  `quantita_disponibile` FLOAT NOT NULL,
  `unita_misura` VARCHAR(20) NOT NULL,
  `soglia_riordino` FLOAT NOT NULL,
  PRIMARY KEY (`id_ingrediente`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Tabella Categoria Pietanza
CREATE TABLE IF NOT EXISTS `categoria_pietanza` (
  `id_categoria` INT NOT NULL AUTO_INCREMENT,
  `nome` VARCHAR(50) NOT NULL,
  PRIMARY KEY (`id_categoria`),
  UNIQUE KEY `unique_nome_categoria` (`nome`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Tabella Pietanza
CREATE TABLE IF NOT EXISTS `pietanza` (
  `id_pietanza` INT NOT NULL AUTO_INCREMENT,
  `nome` VARCHAR(100) NOT NULL,
  `prezzo` DECIMAL(10,2) NOT NULL,
  `id_categoria` INT NOT NULL,
  `disponibile` BOOLEAN NOT NULL DEFAULT TRUE,
  PRIMARY KEY (`id_pietanza`),
  FOREIGN KEY (`id_categoria`) REFERENCES `categoria_pietanza` (`id_categoria`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Tabella Menu (relazione diretta tra ristorante e pietanze)
CREATE TABLE IF NOT EXISTS `menu` (
  `id_ristorante` INT NOT NULL,
  `id_pietanza` INT NOT NULL,
  PRIMARY KEY (`id_ristorante`, `id_pietanza`),
  FOREIGN KEY (`id_ristorante`) REFERENCES `ristorante` (`id_ristorante`) ON DELETE CASCADE,
  FOREIGN KEY (`id_pietanza`) REFERENCES `pietanza` (`id_pietanza`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Tabella Ricetta 
CREATE TABLE IF NOT EXISTS `ricetta` (
  `id_ricetta` INT NOT NULL AUTO_INCREMENT,
  `nome` VARCHAR(100) NOT NULL,
  `descrizione` TEXT NOT NULL,
  `id_pietanza` INT NOT NULL,
  `tempo_preparazione` INT,
  `istruzioni` TEXT,
  PRIMARY KEY (`id_ricetta`),
  FOREIGN KEY (`id_pietanza`) REFERENCES `pietanza` (`id_pietanza`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Tabella Ricetta_Ingrediente (relazione tra ricetta e ingredienti)
CREATE TABLE IF NOT EXISTS `ricetta_ingrediente` (
  `id_ricetta` INT NOT NULL,
  `id_ingrediente` INT NOT NULL,
  `quantita` FLOAT NOT NULL,
  PRIMARY KEY (`id_ricetta`, `id_ingrediente`),
  FOREIGN KEY (`id_ricetta`) REFERENCES `ricetta` (`id_ricetta`) ON DELETE CASCADE,
  FOREIGN KEY (`id_ingrediente`) REFERENCES `ingrediente` (`id_ingrediente`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Tabella Menu Fisso
CREATE TABLE IF NOT EXISTS `menu_fisso` (
  `id_menu` INT NOT NULL AUTO_INCREMENT,
  `nome` VARCHAR(100) NOT NULL,
  `prezzo` DECIMAL(10,2) NOT NULL,
  `descrizione` TEXT,
  PRIMARY KEY (`id_menu`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Tabella Composizione Menu Fisso 
CREATE TABLE IF NOT EXISTS `composizione_menu` (
  `id_menu` INT NOT NULL,
  `id_pietanza` INT NOT NULL,
  PRIMARY KEY (`id_menu`, `id_pietanza`),
  FOREIGN KEY (`id_menu`) REFERENCES `menu_fisso` (`id_menu`) ON DELETE CASCADE,
  FOREIGN KEY (`id_pietanza`) REFERENCES `pietanza` (`id_pietanza`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Tabella Ordine
CREATE TABLE IF NOT EXISTS `ordine` (
  `id_ordine` INT NOT NULL AUTO_INCREMENT,
  `id_tavolo` INT NOT NULL,
  `num_persone` INT NOT NULL,
  `data_ordine` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `stato` ENUM('in_attesa', 'in_preparazione', 'pronto', 'consegnato', 'pagato') NOT NULL DEFAULT 'in_attesa',
  `id_ristorante` INT NOT NULL,
  `costo_totale` DECIMAL(10,2) NOT NULL DEFAULT 0.00,
  PRIMARY KEY (`id_ordine`),
  FOREIGN KEY (`id_tavolo`) REFERENCES `tavolo` (`id_tavolo`),
  FOREIGN KEY (`id_ristorante`) REFERENCES `ristorante` (`id_ristorante`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Tabella Dettaglio Ordine
CREATE TABLE IF NOT EXISTS `dettaglio_ordine_pietanza` (
  `id_dettaglio` INT NOT NULL AUTO_INCREMENT,
  `id_ordine` INT NOT NULL,
  `id_pietanza` INT NOT NULL,
  `quantita` INT NOT NULL DEFAULT 1,
  `parte_di_menu` BOOLEAN NOT NULL DEFAULT FALSE,
  `id_menu` INT DEFAULT NULL,
  PRIMARY KEY (`id_dettaglio`),
  FOREIGN KEY (`id_ordine`) REFERENCES `ordine` (`id_ordine`) ON DELETE CASCADE,
  FOREIGN KEY (`id_pietanza`) REFERENCES `pietanza` (`id_pietanza`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Crea indice per migliorare le ricerche per menu
CREATE INDEX `idx_menu_pietanze` ON `dettaglio_ordine_pietanza` (`id_menu`);

SET FOREIGN_KEY_CHECKS = 1;