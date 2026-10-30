import { useState, type FormEvent } from 'react';
import { useParams } from 'react-router-dom';
import { Aviso, Boton, BotonEnlace, Campo, Chip, Selector, Tarjeta, Td, Th, Titulo } from '../components/ui';
import { useAuth } from '../context/useAuth';
import { useAccion } from '../hooks/useAccion';
import { useCarga } from '../hooks/useCarga';
import {
  anularCobro, ETIQUETA_ESTADO_PAGO, ETIQUETA_METODO, obtenerCobro, registrarPago,
  type Cobro, type MetodoPago,
} from '../services/cobroService';
import { aCentavos, deCentavos, formatearDinero } from '../utils/dinero';

const formatoFechaHora = new Intl.DateTimeFormat('es', { day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' });

const FormularioPago = ({ cobro, alRegistrar }: { cobro: Cobro; alRegistrar: () => void }) => {
  const accion = useAccion();
  const [metodo, setMetodo] = useState<MetodoPago>('efectivo');
  const [monto, setMonto] = useState(deCentavos(cobro.saldo_centavos));
  const [referencia, setReferencia] = useState('');
  const centavos = aCentavos(monto);

  const guardar = async (evento: FormEvent) => {
    evento.preventDefault();
    if (!centavos) return;
    if (await accion.ejecutar(() => registrarPago(cobro.id, metodo, centavos, referencia))) {
      setReferencia('');
      alRegistrar();
    }
  };

  const excede = centavos !== null && centavos > cobro.saldo_centavos;
  return (
    <Tarjeta className="no-imprimir">
      <form onSubmit={guardar} className="space-y-5">
        <h2 className="text-seccion font-semibold">Registrar pago</h2>
        <div className="grid gap-5 md:grid-cols-3">
          <Selector id="pago-metodo" etiqueta="Método" value={metodo} onChange={(e) => setMetodo(e.target.value as MetodoPago)}>
            {(Object.keys(ETIQUETA_METODO) as MetodoPago[]).map((m) => <option key={m} value={m}>{ETIQUETA_METODO[m]}</option>)}
          </Selector>
          <Campo id="pago-monto" etiqueta="Monto" inputMode="decimal" required value={monto} onChange={(e) => setMonto(e.target.value)}
            ayuda={`Saldo pendiente: ${formatearDinero(cobro.saldo_centavos, cobro.moneda)}`} aria-invalid={centavos === null || centavos === 0 || excede} />
          <Campo id="pago-referencia" etiqueta={metodo === 'efectivo' ? 'Referencia (opcional)' : 'Referencia (obligatoria)'} maxLength={100}
            required={metodo !== 'efectivo'} value={referencia} onChange={(e) => setReferencia(e.target.value)} />
        </div>
        {excede && <Aviso>El monto supera el saldo pendiente del cobro.</Aviso>}
        {accion.error && <Aviso>{accion.error}</Aviso>}
        <Boton type="submit" disabled={accion.ocupado || !centavos || excede}>{accion.ocupado ? 'Registrando…' : 'Registrar pago'}</Boton>
      </form>
    </Tarjeta>
  );
};

const PanelAnulacion = ({ cobro, alAnular }: { cobro: Cobro; alAnular: () => void }) => {
  const accion = useAccion();
  const [abierto, setAbierto] = useState(false);
  const [motivo, setMotivo] = useState('');

  const anular = async (evento: FormEvent) => {
    evento.preventDefault();
    if (await accion.ejecutar(() => anularCobro(cobro.id, motivo))) alAnular();
  };

  if (!abierto) return <div className="no-imprimir"><Boton variante="peligro" onClick={() => setAbierto(true)}>Anular cobro</Boton></div>;
  return (
    <Tarjeta className="no-imprimir">
      <form onSubmit={anular} className="space-y-4">
        <h2 className="text-seccion font-semibold">Anular el cobro {cobro.numero}</h2>
        <p className="text-texto-suave">El cobro no se borra: queda registrado como anulado, con el motivo y quién lo anuló. Para corregirlo, genera uno nuevo.</p>
        <Campo id="anulacion-motivo" etiqueta="Motivo de la anulación" required maxLength={300} value={motivo} onChange={(e) => setMotivo(e.target.value)} />
        {accion.error && <Aviso>{accion.error}</Aviso>}
        <div className="flex flex-wrap gap-3">
          <Boton type="submit" variante="peligro" disabled={accion.ocupado || !motivo.trim()}>Confirmar anulación</Boton>
          <Boton variante="secundario" onClick={() => setAbierto(false)}>Volver</Boton>
        </div>
      </form>
    </Tarjeta>
  );
};

const CobroDetallePage = () => {
  const { id } = useParams();
  const { usuario } = useAuth();
  const { datos: cobro, error, recargar } = useCarga(() => obtenerCobro(Number(id)), `cobro-${id}`);
  const esRecepcion = usuario?.rol === 'recepcionista';

  if (error) return <Aviso>{error}</Aviso>;
  if (!cobro) return <p>Cargando…</p>;

  const dinero = (centavos: number) => formatearDinero(centavos, cobro.moneda);
  return (
    <section aria-labelledby="titulo-cobro" className="space-y-6">
      <div className="no-imprimir flex flex-wrap items-center justify-between gap-3">
        <Titulo id="titulo-cobro">Cobro {cobro.numero}</Titulo>
        <div className="flex flex-wrap gap-3">
          <Boton variante="secundario" onClick={() => window.print()}>Imprimir comprobante</Boton>
          <BotonEnlace variante="secundario" to="/cobros">Volver a cobros</BotonEnlace>
        </div>
      </div>

      <Tarjeta className="comprobante">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <p className="text-seccion font-semibold">SIGO-Dental</p>
            <p className="text-texto-suave">Comprobante interno de cobro</p>
          </div>
          <div className="text-right">
            <p className="font-mono text-seccion font-semibold">{cobro.numero}</p>
            <p className="text-etiqueta text-texto-suave">{formatoFechaHora.format(new Date(cobro.creado_en))}</p>
            <div className="mt-2 no-imprimir">
              {cobro.anulado ? <Chip tono="alerta">Anulado</Chip> : <Chip tono={cobro.estado === 'pagado' ? 'acento' : 'neutro'}>{ETIQUETA_ESTADO_PAGO[cobro.estado]}</Chip>}
            </div>
            {cobro.anulado && <p className="hidden font-semibold print:block">COBRO ANULADO</p>}
          </div>
        </div>

        <p className="mt-5">
          <span className="text-texto-suave">Paciente: </span><span className="font-medium">{cobro.paciente_nombre}</span>
          <span className="text-texto-suave"> · Documento </span><span className="font-mono">{cobro.paciente_documento}</span>
        </p>

        <div className="mt-5 overflow-x-auto">
          <table className="w-full min-w-max text-left">
            <caption className="sr-only">Detalle del cobro</caption>
            <thead>
              <tr><Th>Procedimiento</Th><Th>Cantidad</Th><Th>Precio</Th><Th>Importe</Th></tr>
            </thead>
            <tbody>
              {cobro.lineas.map((l) => (
                <tr key={l.id}>
                  <Td>{l.descripcion}</Td>
                  <Td className="font-mono">{l.cantidad}</Td>
                  <Td className="font-mono">{dinero(l.precio_unitario_centavos)}</Td>
                  <Td className="font-mono">{dinero(l.total_centavos)}</Td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <dl className="mt-5 ml-auto max-w-xs space-y-1">
          <div className="flex justify-between"><dt className="text-texto-suave">Subtotal</dt><dd className="font-mono">{dinero(cobro.subtotal_centavos)}</dd></div>
          <div className="flex justify-between"><dt className="text-texto-suave">Descuento ({(cobro.descuento_bp / 100).toFixed(2)} %)</dt><dd className="font-mono">−{dinero(cobro.descuento_centavos)}</dd></div>
          <div className="flex justify-between text-seccion font-semibold"><dt>Total</dt><dd className="font-mono">{dinero(cobro.total_centavos)}</dd></div>
          <div className="flex justify-between"><dt className="text-texto-suave">Pagado</dt><dd className="font-mono">{dinero(cobro.pagado_centavos)}</dd></div>
          <div className="flex justify-between font-medium"><dt>Saldo</dt><dd className="font-mono">{dinero(cobro.saldo_centavos)}</dd></div>
        </dl>

        {cobro.pagos.length > 0 && (
          <>
            <h2 className="mt-6 text-cuerpo font-semibold">Pagos</h2>
            <ul className="mt-2 space-y-1 text-etiqueta">
              {cobro.pagos.map((p) => (
                <li key={p.id}>
                  <span className="font-mono">{formatoFechaHora.format(new Date(p.creado_en))}</span> · {ETIQUETA_METODO[p.metodo]} ·{' '}
                  <span className="font-mono">{dinero(p.monto_centavos)}</span>
                  {p.referencia && <> · ref. <span className="font-mono">{p.referencia}</span></>} · {p.usuario_nombre}
                </li>
              ))}
            </ul>
          </>
        )}

        {cobro.anulacion && (
          <p className="mt-6 text-chip-alerta-texto">
            Anulado por {cobro.anulacion.usuario_nombre} el {formatoFechaHora.format(new Date(cobro.anulacion.creado_en))}: {cobro.anulacion.motivo}
          </p>
        )}
        <p className="mt-6 text-etiqueta text-texto-suave">Atendió: {cobro.creado_por}. Documento interno, sin validez tributaria.</p>
      </Tarjeta>

      {esRecepcion && !cobro.anulado && cobro.saldo_centavos > 0 && <FormularioPago key={cobro.pagado_centavos} cobro={cobro} alRegistrar={recargar} />}
      {esRecepcion && !cobro.anulado && <PanelAnulacion cobro={cobro} alAnular={recargar} />}
    </section>
  );
};

export default CobroDetallePage;
