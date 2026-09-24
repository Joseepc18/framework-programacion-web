// src/data/productos.ts
// Datos de ejemplo (mock) del catálogo.
// NOTA: en la Unidad 2 estos datos se reemplazarán por los registros reales
// de PostgreSQL consultados a través de GET /api/productos.

export interface Producto {
  id: number;
  nombre: string;
  descripcion: string;
  precio: number;
  categoria: string;
  img: string;        // Imagen principal (portada)
  galeria: string[];  // Imágenes adicionales para la galería/lightbox
}

// Lista de categorías derivada de los productos (se usa para los filtros)
export const categorias = [
  "Serum",
  "Crema",
  "Tónico",
  "Mascarilla",
  "Kit",
];

export const productosMock: Producto[] = [
  {
    id: 1,
    nombre: "Serum Revitalizante",
    descripcion:
      "Serum concentrado con vitamina C y ácido hialurónico. Ilumina la piel, reduce manchas y aporta hidratación profunda desde la primera aplicación.",
    precio: 45.0,
    categoria: "Serum",
    img: "https://picsum.photos/seed/serum/600",
    galeria: [
      "https://picsum.photos/seed/serum1/600",
      "https://picsum.photos/seed/serum2/600",
      "https://picsum.photos/seed/serum3/600",
    ],
  },
  {
    id: 2,
    nombre: "Crema Hidratante Pro",
    descripcion:
      "Crema facial de textura ligera con niacinamida y manteca de karité. Hidratación de 24 horas y barrera cutánea fortalecida para todo tipo de piel.",
    precio: 32.5,
    categoria: "Crema",
    img: "https://picsum.photos/seed/crema/600",
    galeria: [
      "https://picsum.photos/seed/crema1/600",
      "https://picsum.photos/seed/crema2/600",
      "https://picsum.photos/seed/crema3/600",
    ],
  },
  {
    id: 3,
    nombre: "Tónico Purificante",
    descripcion:
      "Tónico sin alcohol con hamamelis y agua de rosas. Limpia poros, equilibra el pH y prepara la piel para el resto de la rutina diaria.",
    precio: 28.0,
    categoria: "Tónico",
    img: "https://picsum.photos/seed/tonico/600",
    galeria: [
      "https://picsum.photos/seed/tonico1/600",
      "https://picsum.photos/seed/tonico2/600",
      "https://picsum.photos/seed/tonico3/600",
    ],
  },
  {
    id: 4,
    nombre: "Mascarilla Nocturna",
    descripcion:
      "Mascarilla de noche con colágeno y vitamina E. Repara la piel mientras duermes y recupera la luminosidad al despertar.",
    precio: 50.0,
    categoria: "Mascarilla",
    img: "https://picsum.photos/seed/mascarilla/600",
    galeria: [
      "https://picsum.photos/seed/mascarilla1/600",
      "https://picsum.photos/seed/mascarilla2/600",
      "https://picsum.photos/seed/mascarilla3/600",
    ],
  },
  {
    id: 5,
    nombre: "Serum Anti-Edad Retinol",
    descripcion:
      "Serum de retinol encapsulado que suaviza líneas de expresión y mejora la firmeza. Uso nocturno recomendado.",
    precio: 58.0,
    categoria: "Serum",
    img: "https://picsum.photos/seed/retinol/600",
    galeria: [
      "https://picsum.photos/seed/retinol1/600",
      "https://picsum.photos/seed/retinol2/600",
      "https://picsum.photos/seed/retinol3/600",
    ],
  },
  {
    id: 6,
    nombre: "Crema Contorno de Ojos",
    descripcion:
      "Contorno de ojos con cafeína y péptidos. Reduce bolsas y ojeras, hidrata la zona más delicada del rostro.",
    precio: 26.0,
    categoria: "Crema",
    img: "https://picsum.photos/seed/ojos/600",
    galeria: [
      "https://picsum.photos/seed/ojos1/600",
      "https://picsum.photos/seed/ojos2/600",
      "https://picsum.photos/seed/ojos3/600",
    ],
  },
  {
    id: 7,
    nombre: "Tónico Exfoliante AHA-BHA",
    descripcion:
      "Exfoliación química suave con ácidos AHA y BHA. Renueva la textura de la piel y desobstruye los poros.",
    precio: 34.0,
    categoria: "Tónico",
    img: "https://picsum.photos/seed/exfoliante/600",
    galeria: [
      "https://picsum.photos/seed/exfoliante1/600",
      "https://picsum.photos/seed/exfoliante2/600",
      "https://picsum.photos/seed/exfoliante3/600",
    ],
  },
  {
    id: 8,
    nombre: "Kit Rutina Completa",
    descripcion:
      "Kit integral con serum, crema, tónico y mascarilla. La rutina perfecta para comenzar: limpieza, tratamiento e hidratación.",
    precio: 129.0,
    categoria: "Kit",
    img: "https://picsum.photos/seed/kit/600",
    galeria: [
      "https://picsum.photos/seed/kit1/600",
      "https://picsum.photos/seed/kit2/600",
      "https://picsum.photos/seed/kit3/600",
    ],
  },
];