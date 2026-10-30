// Numeración dental y catálogo de hallazgos del odontograma (RF-03). Internamente todo se guarda en FDI; la numeración
// Universal solo cambia lo que se muestra.

export type Denticion = 'permanente' | 'temporal' | 'mixta';
export type Numeracion = 'fdi' | 'universal';
export type Superficie = 'mesial' | 'distal' | 'oclusal' | 'vestibular' | 'lingual';
export type TipoHallazgo =
  | 'caries' | 'restauracion' | 'endodoncia' | 'implante' | 'protesis_fija' | 'protesis_removible' | 'exodoncia' | 'sellante' | 'ausente';
export type EstadoHallazgo = 'requerido' | 'existente' | 'realizado';

// Filas del odontograma de izquierda a derecha, como se ve al paciente de frente: cuadrante derecho del paciente a la izquierda.
export const ARCADAS = {
  permanente: {
    superior: [18, 17, 16, 15, 14, 13, 12, 11, 21, 22, 23, 24, 25, 26, 27, 28],
    inferior: [48, 47, 46, 45, 44, 43, 42, 41, 31, 32, 33, 34, 35, 36, 37, 38],
  },
  temporal: {
    superior: [55, 54, 53, 52, 51, 61, 62, 63, 64, 65],
    inferior: [85, 84, 83, 82, 81, 71, 72, 73, 74, 75],
  },
} as const;

export const TODAS_LAS_PIEZAS: readonly number[] = [
  ...ARCADAS.permanente.superior, ...ARCADAS.permanente.inferior, ...ARCADAS.temporal.superior, ...ARCADAS.temporal.inferior,
];

export const esPermanente = (pieza: number) => pieza >= 11 && pieza <= 48;
export const esSuperior = (pieza: number) => [1, 2, 5, 6].includes(Math.floor(pieza / 10));
// Los cuadrantes 1, 4, 5 y 8 son el lado derecho del paciente: el centro de la boca queda a la derecha del diente.
export const centroALaDerecha = (pieza: number) => [1, 4, 5, 8].includes(Math.floor(pieza / 10));

// Pasa una pieza FDI a Universal: 1 a 32 en permanentes y A a T en temporales (misma tabla que el servidor).
export function aUniversal(pieza: number): string {
  const cuadrante = Math.floor(pieza / 10);
  const posicion = pieza % 10;
  if (esPermanente(pieza)) {
    const n = [0, 9 - posicion, 8 + posicion, 25 - posicion, 24 + posicion][cuadrante];
    return String(n);
  }
  const n = [0, 0, 0, 0, 0, 5 - posicion, 4 + posicion, 15 - posicion, 14 + posicion][cuadrante];
  return String.fromCharCode(65 + n);
}

export const numeroDePieza = (pieza: number, numeracion: Numeracion) => (numeracion === 'fdi' ? String(pieza) : aUniversal(pieza));

const NOMBRES_DIENTE = ['', 'incisivo central', 'incisivo lateral', 'canino', 'primer premolar', 'segundo premolar', 'primer molar', 'segundo molar', 'tercer molar'];

export function nombreDePieza(pieza: number): string {
  const posicion = pieza % 10;
  if (esPermanente(pieza)) return NOMBRES_DIENTE[posicion];
  return ['', 'incisivo central', 'incisivo lateral', 'canino', 'primer molar', 'segundo molar'][posicion];
}

// La cara de masticación se llama incisal en incisivos y caninos; la lingual se llama palatina en el maxilar superior.
export function etiquetaSuperficie(superficie: Superficie, pieza: number): string {
  if (superficie === 'oclusal') return pieza % 10 <= 3 ? 'Incisal' : 'Oclusal';
  if (superficie === 'lingual') return esSuperior(pieza) ? 'Palatina' : 'Lingual';
  return { mesial: 'Mesial', distal: 'Distal', vestibular: 'Vestibular' }[superficie];
}

export const SUPERFICIES: readonly Superficie[] = ['mesial', 'distal', 'oclusal', 'vestibular', 'lingual'];

export interface MetaTipo {
  etiqueta: string;
  // Se marca por superficie (true) o afecta a la pieza completa (false).
  porSuperficie: boolean;
  estados: readonly EstadoHallazgo[];
  // Clases fijas (Tailwind las lee del código): color con tokens Aurora.
  relleno: string;
  trazo: string;
  // Texto corto que acompaña al color en la leyenda.
  forma: string;
}

export const TIPOS: Record<TipoHallazgo, MetaTipo> = {
  caries: { etiqueta: 'Caries', porSuperficie: true, estados: ['requerido'], relleno: 'fill-odonto-caries', trazo: 'stroke-odonto-caries-borde', forma: 'superficie en coral' },
  restauracion: { etiqueta: 'Restauración', porSuperficie: true, estados: ['existente', 'requerido', 'realizado'], relleno: 'fill-odonto-restauracion', trazo: 'stroke-odonto-restauracion-borde', forma: 'superficie en azul' },
  sellante: { etiqueta: 'Sellante', porSuperficie: true, estados: ['requerido', 'existente', 'realizado'], relleno: 'fill-odonto-sellante', trazo: 'stroke-odonto-sellante-borde', forma: 'superficie en turquesa' },
  endodoncia: { etiqueta: 'Endodoncia', porSuperficie: false, estados: ['requerido', 'existente', 'realizado'], relleno: 'fill-odonto-endodoncia', trazo: 'stroke-odonto-endodoncia-borde', forma: 'triángulo ámbar' },
  implante: { etiqueta: 'Implante', porSuperficie: false, estados: ['requerido', 'existente', 'realizado'], relleno: 'fill-odonto-implante', trazo: 'stroke-odonto-implante-borde', forma: 'tornillo violeta' },
  protesis_fija: { etiqueta: 'Prótesis fija', porSuperficie: false, estados: ['requerido', 'existente', 'realizado'], relleno: 'fill-odonto-protesis-fija', trazo: 'stroke-odonto-protesis-fija', forma: 'marco continuo' },
  protesis_removible: { etiqueta: 'Prótesis removible', porSuperficie: false, estados: ['requerido', 'existente', 'realizado'], relleno: 'fill-odonto-protesis-removible', trazo: 'stroke-odonto-protesis-removible', forma: 'marco de rayas' },
  exodoncia: { etiqueta: 'Exodoncia', porSuperficie: false, estados: ['requerido', 'realizado'], relleno: 'fill-odonto-exodoncia', trazo: 'stroke-odonto-exodoncia', forma: 'cruz' },
  ausente: { etiqueta: 'Diente ausente', porSuperficie: false, estados: ['existente'], relleno: 'fill-transparent', trazo: 'stroke-odonto-ausente-borde', forma: 'contorno punteado' },
};

export const ORDEN_TIPOS = Object.keys(TIPOS) as TipoHallazgo[];

export const ETIQUETA_ESTADO: Record<EstadoHallazgo, string> = {
  requerido: 'Requerido',
  existente: 'Existente',
  realizado: 'Realizado',
};

// Qué significa cada estado, para la leyenda y la ayuda del formulario.
export const AYUDA_ESTADO: Record<EstadoHallazgo, string> = {
  requerido: 'por hacer; se puede corregir o retirar',
  existente: 'el paciente ya lo traía; se puede corregir o retirar',
  realizado: 'hecho en la clínica; no se corrige ni se retira (RN-01)',
};

// Geometría de los cinco polígonos de un diente en un cuadro de 40 x 40: cuatro trapecios y el centro.
export const POLIGONOS = {
  arriba: '0,0 40,0 28,12 12,12',
  abajo: '0,40 40,40 28,28 12,28',
  izquierda: '0,0 12,12 12,28 0,40',
  derecha: '40,0 28,12 28,28 40,40',
  centro: '12,12 28,12 28,28 12,28',
} as const;

export type Zona = keyof typeof POLIGONOS;

// Qué zona del dibujo corresponde a cada superficie según el diente: el vestibular mira hacia afuera (arriba en el
// maxilar superior, abajo en el inferior) y el mesial hacia el centro de la boca.
export function zonaDeSuperficie(superficie: Superficie, pieza: number): Zona {
  const superior = esSuperior(pieza);
  switch (superficie) {
    case 'oclusal':
      return 'centro';
    case 'vestibular':
      return superior ? 'arriba' : 'abajo';
    case 'lingual':
      return superior ? 'abajo' : 'arriba';
    case 'mesial':
      return centroALaDerecha(pieza) ? 'derecha' : 'izquierda';
    default:
      return centroALaDerecha(pieza) ? 'izquierda' : 'derecha';
  }
}

// Dentición que se muestra por defecto según la edad: temporal hasta los 5 años, mixta de 6 a 12 y permanente después.
export const denticionPorEdad = (edad: number): Denticion => (edad <= 5 ? 'temporal' : edad <= 12 ? 'mixta' : 'permanente');
