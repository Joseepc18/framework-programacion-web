-- Script de inicialización de la base de datos multicatalogo.
-- PostgreSQL lo ejecuta automáticamente la primera vez que se crea el contenedor
-- (está montado en /docker-entrypoint-initdb.d desde docker-compose.yml).

-- Tabla de cuentas del sistema. La contraseña se guarda en texto plano de forma temporal.
CREATE TABLE usuarios (
    id       SERIAL PRIMARY KEY,
    email    VARCHAR(150) NOT NULL UNIQUE,
    password VARCHAR(100) NOT NULL,
    rol      VARCHAR(20)  NOT NULL CHECK (rol IN ('admin', 'cliente'))
);

-- Tabla del catálogo. La galería se guarda como un arreglo de URLs (TEXT[]).
CREATE TABLE productos (
    id          SERIAL PRIMARY KEY,
    nombre      VARCHAR(150)   NOT NULL,
    descripcion TEXT           NOT NULL DEFAULT '',
    precio      NUMERIC(10, 2) NOT NULL CHECK (precio >= 0),
    categoria   VARCHAR(50)    NOT NULL,
    img         TEXT           NOT NULL,
    galeria     TEXT[]         NOT NULL DEFAULT '{}'
);

-- Tabla de la red multinivel. parent_id apunta a la misma tabla para formar el árbol;
-- la raíz (el usuario autenticado) es la única fila con parent_id NULL.
CREATE TABLE referidos (
    id        SERIAL PRIMARY KEY,
    nombre    VARCHAR(150)   NOT NULL,
    nivel     INTEGER        NOT NULL CHECK (nivel >= 0),
    ventas    NUMERIC(10, 2) NOT NULL DEFAULT 0,
    parent_id INTEGER        REFERENCES referidos (id) ON DELETE CASCADE
);

-- Datos semilla: los mismos que existían en memoria en repository.go.
INSERT INTO usuarios (id, email, password, rol) VALUES
    (1, 'admin@upse.edu.ec', '123456', 'admin'),
    (2, 'cliente@upse.edu.ec', '123456', 'cliente');

-- Datos semilla del catálogo: los 8 productos que usa el frontend.
INSERT INTO productos (id, nombre, descripcion, precio, categoria, img, galeria) VALUES
    (1, 'Serum Revitalizante',
     'Serum concentrado con vitamina C y ácido hialurónico. Ilumina la piel, reduce manchas y aporta hidratación profunda desde la primera aplicación.',
     45.00, 'Serum', 'https://picsum.photos/seed/serum/600',
     ARRAY['https://picsum.photos/seed/serum1/600', 'https://picsum.photos/seed/serum2/600', 'https://picsum.photos/seed/serum3/600']),
    (2, 'Crema Hidratante Pro',
     'Crema facial de textura ligera con niacinamida y manteca de karité. Hidratación de 24 horas y barrera cutánea fortalecida para todo tipo de piel.',
     32.50, 'Crema', 'https://picsum.photos/seed/crema/600',
     ARRAY['https://picsum.photos/seed/crema1/600', 'https://picsum.photos/seed/crema2/600', 'https://picsum.photos/seed/crema3/600']),
    (3, 'Tónico Purificante',
     'Tónico sin alcohol con hamamelis y agua de rosas. Limpia poros, equilibra el pH y prepara la piel para el resto de la rutina diaria.',
     28.00, 'Tónico', 'https://picsum.photos/seed/tonico/600',
     ARRAY['https://picsum.photos/seed/tonico1/600', 'https://picsum.photos/seed/tonico2/600', 'https://picsum.photos/seed/tonico3/600']),
    (4, 'Mascarilla Nocturna',
     'Mascarilla de noche con colágeno y vitamina E. Repara la piel mientras duermes y recupera la luminosidad al despertar.',
     50.00, 'Mascarilla', 'https://picsum.photos/seed/mascarilla/600',
     ARRAY['https://picsum.photos/seed/mascarilla1/600', 'https://picsum.photos/seed/mascarilla2/600', 'https://picsum.photos/seed/mascarilla3/600']),
    (5, 'Serum Anti-Edad Retinol',
     'Serum de retinol encapsulado que suaviza líneas de expresión y mejora la firmeza. Uso nocturno recomendado.',
     58.00, 'Serum', 'https://picsum.photos/seed/retinol/600',
     ARRAY['https://picsum.photos/seed/retinol1/600', 'https://picsum.photos/seed/retinol2/600', 'https://picsum.photos/seed/retinol3/600']),
    (6, 'Crema Contorno de Ojos',
     'Contorno de ojos con cafeína y péptidos. Reduce bolsas y ojeras, hidrata la zona más delicada del rostro.',
     26.00, 'Crema', 'https://picsum.photos/seed/ojos/600',
     ARRAY['https://picsum.photos/seed/ojos1/600', 'https://picsum.photos/seed/ojos2/600', 'https://picsum.photos/seed/ojos3/600']),
    (7, 'Tónico Exfoliante AHA-BHA',
     'Exfoliación química suave con ácidos AHA y BHA. Renueva la textura de la piel y desobstruye los poros.',
     34.00, 'Tónico', 'https://picsum.photos/seed/exfoliante/600',
     ARRAY['https://picsum.photos/seed/exfoliante1/600', 'https://picsum.photos/seed/exfoliante2/600', 'https://picsum.photos/seed/exfoliante3/600']),
    (8, 'Kit Rutina Completa',
     'Kit integral con serum, crema, tónico y mascarilla. La rutina perfecta para comenzar: limpieza, tratamiento e hidratación.',
     129.00, 'Kit', 'https://picsum.photos/seed/kit/600',
     ARRAY['https://picsum.photos/seed/kit1/600', 'https://picsum.photos/seed/kit2/600', 'https://picsum.photos/seed/kit3/600']);

-- Datos semilla de la red. Se insertan de arriba hacia abajo para respetar la clave foránea.
INSERT INTO referidos (id, nombre, nivel, ventas, parent_id) VALUES
    (0, 'Tú',            0, 2400, NULL),
    (1, 'Ana García',    1, 1200, 0),
    (2, 'Luis Poveda',   1,  850, 0),
    (3, 'Marta Sánchez', 1,  430, 0),
    (4, 'Carlos Ruiz',   2,  500, 1),
    (5, 'Sofía León',    2,  430, 1),
    (6, 'Marco Díaz',    2,  380, 2),
    (7, 'Diana Paz',     3,  300, 4);

-- Como los IDs se insertaron a mano, ajustamos las secuencias para que los próximos INSERT no choquen.
SELECT setval('usuarios_id_seq',  (SELECT MAX(id) FROM usuarios));
SELECT setval('productos_id_seq', (SELECT MAX(id) FROM productos));
SELECT setval('referidos_id_seq', (SELECT MAX(id) FROM referidos));
