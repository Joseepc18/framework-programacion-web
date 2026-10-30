// Acceso a la API de cobros y comprobantes. El dinero viaja en centavos y los importes los calcula el servidor.
import { apiFetch } from './http';

export type MetodoPago = 'efectivo' | 'tarjeta' | 'transferencia';
export type EstadoPago = 'pendiente' | 'parcial' | 'pagado';

export interface CobroResumen {
  id: number;
  numero: string;
  paciente_id: number;
  paciente_nombre: string;
  total_centavos: number;
  pagado_centavos: number;
  saldo_centavos: number;
  estado: EstadoPago;
  anulado: boolean;
  creado_en: string;
}

export interface CobroLinea {
  id: number;
  procedimiento_id: number;
  descripcion: string;
  cantidad: number;
  precio_unitario_centavos: number;
  total_centavos: number;
}

export interface Pago {
  id: number;
  metodo: MetodoPago;
  monto_centavos: number;
  referencia: string;
  usuario_nombre: string;
  creado_en: string;
}

export interface Cobro extends CobroResumen {
  cita_id: number | null;
  paciente_documento: string;
  subtotal_centavos: number;
  descuento_bp: number;
  descuento_centavos: number;
  moneda: string;
  creado_por: string;
  lineas: CobroLinea[];
  pagos: Pago[];
  anulacion: { motivo: string; usuario_nombre: string; creado_en: string } | null;
}

export interface LineaPedida {
  procedimiento_id: number;
  cantidad: number;
}

export interface DatosCobro {
  paciente_id?: number;
  cita_id?: number;
  descuento_bp: number;
  lineas: LineaPedida[];
}

export const ETIQUETA_METODO: Record<MetodoPago, string> = { efectivo: 'Efectivo', tarjeta: 'Tarjeta', transferencia: 'Transferencia' };
export const ETIQUETA_ESTADO_PAGO: Record<EstadoPago, string> = { pendiente: 'Pendiente', parcial: 'Pagado parcial', pagado: 'Pagado' };

export function listarCobros(busqueda: string) {
  const texto = busqueda.trim();
  return apiFetch<CobroResumen[]>(`/api/cobros${texto ? `?q=${encodeURIComponent(texto)}` : ''}`);
}

export const obtenerCobro = (id: number) => apiFetch<Cobro>(`/api/cobros/${id}`);

export const crearCobro = (datos: DatosCobro) => apiFetch<Cobro>('/api/cobros', { method: 'POST', body: JSON.stringify(datos) });

export const registrarPago = (id: number, metodo: MetodoPago, montoCentavos: number, referencia: string) =>
  apiFetch<Cobro>(`/api/cobros/${id}/pagos`, { method: 'POST', body: JSON.stringify({ metodo, monto_centavos: montoCentavos, referencia }) });

export const anularCobro = (id: number, motivo: string) =>
  apiFetch<Cobro>(`/api/cobros/${id}/anular`, { method: 'POST', body: JSON.stringify({ motivo }) });
