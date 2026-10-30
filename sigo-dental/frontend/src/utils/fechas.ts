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
