// src/data/red.ts
// Datos de ejemplo (mock) de la red multinivel de referidos.
// Incluye la estructura jerárquica (árbol) y funciones puras para calcular
// comisiones, totales y el nivel alcanzado por el usuario.

export interface Referido {
  id: number;
  nombre: string;
  nivel: number;     // 1 = directo, 2 = indirecto, 3 = tercer nivel
  ventas: number;    // ventas mensuales en dólares
  hijos?: Referido[]; // referidos de un nivel más abajo
}

// Porcentaje de comisión que recibe el usuario por las ventas de cada nivel
export const TASA_COMISION: Record<number, number> = {
  1: 0.10, // 10 % de las ventas de nivel 1
  2: 0.05, // 5 % de las ventas de nivel 2
  3: 0.02, // 2 % de las ventas de nivel 3
};

// Raíz del árbol: el usuario autenticado (nivel 0) con sus referidos
export const redInicial: Referido = {
  id: 0,
  nombre: "Tú",
  nivel: 0,
  ventas: 2400,
  hijos: [
    {
      id: 1,
      nombre: "Ana García",
      nivel: 1,
      ventas: 1200,
      hijos: [
        {
          id: 4,
          nombre: "Carlos Ruiz",
          nivel: 2,
          ventas: 500,
          hijos: [
            { id: 7, nombre: "Diana Paz", nivel: 3, ventas: 300 },
          ],
        },
        { id: 5, nombre: "Sofía León", nivel: 2, ventas: 430 },
      ],
    },
    {
      id: 2,
      nombre: "Luis Poveda",
      nivel: 1,
      ventas: 850,
      hijos: [
        { id: 6, nombre: "Marco Díaz", nivel: 2, ventas: 380 },
      ],
    },
    { id: 3, nombre: "Marta Sánchez", nivel: 1, ventas: 430 },
  ],
};

// Total de referidos de la red (sin contar al usuario raíz)
export const contarRed = (raiz: Referido): number => {
  const directos = raiz.hijos ?? [];
  return directos.reduce((total, hijo) => total + 1 + contarRed(hijo), 0);
};

// Suma de ventas de toda la red (excluye las ventas propias del usuario raíz)
export const sumarVentasRed = (raiz: Referido): number => {
  const directos = raiz.hijos ?? [];
  return directos.reduce((total, hijo) => total + hijo.ventas + sumarVentasRed(hijo), 0);
};

// Comisión que recibe el usuario por las ventas de cada referido (ventas * tasa del nivel)
export const comisionDeReferido = (ref: Referido): number =>
  ref.ventas * (TASA_COMISION[ref.nivel] ?? 0);

// Comisiones totales del mes acumuladas sobre toda la red
export const sumarComisiones = (raiz: Referido): number => {
  const directos = raiz.hijos ?? [];
  return directos.reduce(
    (total, hijo) => total + comisionDeReferido(hijo) + sumarComisiones(hijo),
    0
  );
};

// Nivel alcanzado por el usuario según la cantidad de referidos directos
export const nivelAlcanzado = (referidosDirectos: number): string => {
  if (referidosDirectos >= 6) return "Diamante";
  if (referidosDirectos >= 4) return "Oro";
  if (referidosDirectos >= 2) return "Plata";
  return "Bronce";
};