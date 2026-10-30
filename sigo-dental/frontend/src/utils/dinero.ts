// El dinero viaja siempre en centavos enteros; solo se convierte al mostrarlo o al leerlo de un formulario.
// La moneda es un parámetro del sistema (decisión de país pendiente): por defecto dólares.
const FORMATOS = new Map<string, Intl.NumberFormat>();

export function formatearDinero(centavos: number, moneda = 'USD') {
  let formato = FORMATOS.get(moneda);
  if (!formato) {
    formato = new Intl.NumberFormat('es-EC', { style: 'currency', currency: moneda });
    FORMATOS.set(moneda, formato);
  }
  return formato.format(centavos / 100);
}

// Convierte "12", "12.5" o "12,50" a centavos; devuelve null si no es un importe válido.
export function aCentavos(texto: string): number | null {
  const limpio = texto.trim().replace(',', '.');
  if (!/^\d+(\.\d{1,2})?$/.test(limpio)) return null;
  const [enteros, decimales = ''] = limpio.split('.');
  return Number(enteros) * 100 + Number(decimales.padEnd(2, '0'));
}

// Para rellenar un campo de importe a partir de centavos.
export const deCentavos = (centavos: number) => (centavos / 100).toFixed(2);
