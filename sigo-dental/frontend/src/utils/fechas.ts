// Fechas de la interfaz. Las fechas de calendario viajan como AAAA-MM-DD y las horas como ISO-8601.

// Años cumplidos a la fecha dada (hoy por defecto); null si el texto no es una fecha AAAA-MM-DD válida.
export function edadDe(fecha: string, hoy = new Date()): number | null {
  const partes = /^(\d{4})-(\d{2})-(\d{2})$/.exec(fecha);
  if (!partes) return null;
  const [anio, mes, dia] = partes.slice(1).map(Number);
  let edad = hoy.getFullYear() - anio;
  if (hoy.getMonth() + 1 < mes || (hoy.getMonth() + 1 === mes && hoy.getDate() < dia)) edad--;
  return edad;
}

// Fecha local de hoy en formato AAAA-MM-DD (sin pasar por UTC, que puede dar el día siguiente).
export const hoyISO = (hoy = new Date()) => aISOFecha(hoy);

export function aISOFecha(fecha: Date): string {
  const mes = String(fecha.getMonth() + 1).padStart(2, '0');
  const dia = String(fecha.getDate()).padStart(2, '0');
  return `${fecha.getFullYear()}-${mes}-${dia}`;
}

// --- Agenda: las citas viajan en ISO-8601 y se muestran en la hora local del navegador. ---

export const sumarDias = (fecha: Date, dias: number) => {
  const nueva = new Date(fecha);
  nueva.setDate(nueva.getDate() + dias);
  return nueva;
};

// Medianoche local del día de la fecha dada.
export const inicioDelDia = (fecha: Date) => new Date(fecha.getFullYear(), fecha.getMonth(), fecha.getDate());

// Lunes de la semana de la fecha dada, a medianoche local.
export function inicioDeSemana(fecha: Date): Date {
  const dia = inicioDelDia(fecha);
  return sumarDias(dia, -((dia.getDay() + 6) % 7));
}

// Interpreta AAAA-MM-DD como medianoche local (new Date('AAAA-MM-DD') lo tomaría como UTC).
export function desdeISOFecha(texto: string): Date {
  const [anio, mes, dia] = texto.split('-').map(Number);
  return new Date(anio, mes - 1, dia);
}

export const horaLocal = (fecha: Date) =>
  `${String(fecha.getHours()).padStart(2, '0')}:${String(fecha.getMinutes()).padStart(2, '0')}`;

// Combina una fecha AAAA-MM-DD y una hora HH:MM locales en un instante.
export function combinarFechaHora(fecha: string, hora: string): Date {
  const [h, m] = hora.split(':').map(Number);
  const base = desdeISOFecha(fecha);
  base.setHours(h, m, 0, 0);
  return base;
}

// "HH:MM" a minutos desde la medianoche.
export const minutosDe = (hhmm: string) => {
  const [h, m] = hhmm.split(':').map(Number);
  return h * 60 + m;
};

const DIAS = ['domingo', 'lunes', 'martes', 'miércoles', 'jueves', 'viernes', 'sábado'];
export const nombreDelDia = (fecha: Date) => DIAS[fecha.getDay()];
