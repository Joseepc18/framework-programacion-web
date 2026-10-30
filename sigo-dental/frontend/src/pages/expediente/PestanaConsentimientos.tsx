import { useState, type FormEvent } from 'react';
import LienzoFirma from '../../components/LienzoFirma';
import { Aviso, Boton, Campo, Chip, Selector, Tarjeta } from '../../components/ui';
import { useAuth } from '../../context/useAuth';
import { useAccion } from '../../hooks/useAccion';
import { mensajeDeApi, useCarga } from '../../hooks/useCarga';
import { listarProcedimientos } from '../../services/catalogoService';
import {
  crearConsentimiento, enviarCodigo, firmarConConsentimientoOtp, firmarManuscrito, listarCodigosEnviados, listarConsentimientos,
  revocarConsentimiento, type Consentimiento,
} from '../../services/consentimientoService';
import { descargarArchivo } from '../../services/expedienteService';
import { abrirBlob } from '../../utils/archivos';
import { fechaHoraLocal } from '../../utils/fechas';

type Modo = 'manuscrita' | 'otp' | 'revocar' | null;

const FilaConsentimiento = ({ c, pacienteId, alCambiar, puedeRevocar }: { c: Consentimiento; pacienteId: number; alCambiar: () => void; puedeRevocar: boolean }) => {
  const accion = useAccion();
  const [modo, setModo] = useState<Modo>(null);
  const [codigo, setCodigo] = useState('');
  const [aviso, setAviso] = useState('');
  const [motivo, setMotivo] = useState('');
  const [errorPdf, setErrorPdf] = useState('');
  const [texto, setTexto] = useState(false);

  const enviar = async () => {
    setAviso('');
    let destino = '';
    if (await accion.ejecutar(async () => { destino = (await enviarCodigo(pacienteId, c.id)).enviado_a; })) {
      setAviso(`Código enviado a ${destino}. Vence en 10 minutos.`);
    }
  };
  const terminar = () => { setModo(null); alCambiar(); };

  const abrirPdf = async () => {
    setErrorPdf('');
    try {
      abrirBlob(await descargarArchivo(c.archivo_id!), `consentimiento-${c.id}.pdf`);
    } catch (fallo) {
      setErrorPdf(mensajeDeApi(fallo));
    }
  };

  const etiquetaEstado = c.estado === 'pendiente' ? 'Pendiente de firma' : c.tipo === 'revocacion' ? 'Revocación firmada' : c.revocado ? 'Revocado' : c.vigente ? 'Vigente' : 'Firmado (vencido)';
  const tono = c.estado === 'pendiente' ? 'neutro' : c.tipo === 'consentimiento' && c.vigente ? 'acento' : c.revocado ? 'alerta' : 'neutro';

  return (
    <li className="space-y-3 rounded-control border border-vidrio-borde p-4">
      <div className="flex flex-wrap items-center gap-3">
        <span className="font-medium">{c.tipo === 'revocacion' ? 'Revocación · ' : ''}{c.procedimiento_nombre}</span>
        <Chip tono={tono}>{etiquetaEstado}</Chip>
        <span className="text-etiqueta text-texto-suave">
          Firma {c.firmante_tipo === 'tutor' ? 'del tutor' : 'del paciente'}: {c.firmante_nombre}
          {c.firmado_en && <> · {c.modalidad === 'otp' ? 'código OTP' : 'firma manuscrita'} · {fechaHoraLocal(c.firmado_en)}</>}
        </span>
      </div>
      <button type="button" className="text-etiqueta text-acento underline" onClick={() => setTexto(!texto)}>{texto ? 'Ocultar el texto' : 'Ver el texto'}</button>
      {texto && <p className="whitespace-pre-wrap rounded-control bg-chip-neutro-fondo p-3 text-etiqueta">{c.texto}</p>}
      {c.estado === 'firmado' && <p className="font-mono text-etiqueta text-texto-suave">SHA-256 del texto {c.huella_texto.slice(0, 16)}…</p>}
      {errorPdf && <Aviso>{errorPdf}</Aviso>}
      {accion.error && <Aviso>{accion.error}</Aviso>}
      {aviso && <Aviso tipo="info">{aviso}</Aviso>}

      {c.estado === 'firmado' && c.archivo_id && (
        <div className="flex flex-wrap gap-3">
          <Boton variante="secundario" onClick={abrirPdf}>Abrir el PDF firmado</Boton>
          {puedeRevocar && c.tipo === 'consentimiento' && !c.revocado && modo !== 'revocar' && <Boton variante="secundario" onClick={() => setModo('revocar')}>Revocar</Boton>}
        </div>
      )}

      {c.estado === 'pendiente' && modo === null && (
        <div className="flex flex-wrap gap-3">
          <Boton onClick={() => setModo('manuscrita')}>Firmar en pantalla</Boton>
          <Boton variante="secundario" onClick={() => { setModo('otp'); void enviar(); }}>Enviar código al {c.firmante_tipo === 'tutor' ? 'tutor' : 'paciente'}</Boton>
        </div>
      )}
      {modo === 'manuscrita' && (
        <div className="space-y-3">
          <LienzoFirma ocupado={accion.ocupado} alConfirmar={async (firma) => { if (await accion.ejecutar(() => firmarManuscrito(pacienteId, c.id, firma))) terminar(); }} />
          <Boton variante="secundario" onClick={() => setModo(null)}>Cancelar</Boton>
        </div>
      )}
      {modo === 'otp' && (
        <form className="space-y-3" onSubmit={async (e: FormEvent) => { e.preventDefault(); if (await accion.ejecutar(() => firmarConConsentimientoOtp(pacienteId, c.id, codigo))) terminar(); }}>
          <Campo id={`otp-${c.id}`} etiqueta="Código recibido por el paciente" inputMode="numeric" maxLength={6} required autoComplete="one-time-code" value={codigo}
            onChange={(e) => setCodigo(e.target.value.replace(/\D/g, ''))} ayuda="Seis dígitos. Vence a los 10 minutos y admite 3 intentos." />
          <div className="flex flex-wrap gap-3">
            <Boton type="submit" disabled={accion.ocupado || codigo.length !== 6}>Validar y firmar</Boton>
            <Boton variante="secundario" disabled={accion.ocupado} onClick={() => void enviar()}>Reenviar código</Boton>
            <Boton variante="secundario" onClick={() => setModo(null)}>Cancelar</Boton>
          </div>
        </form>
      )}
      {modo === 'revocar' && (
        <form className="space-y-3" onSubmit={async (e: FormEvent) => { e.preventDefault(); if (await accion.ejecutar(() => revocarConsentimiento(pacienteId, c.id, motivo))) terminar(); }}>
          <Campo id={`revocar-${c.id}`} etiqueta="Motivo de la revocación" required maxLength={300} value={motivo} onChange={(e) => setMotivo(e.target.value)}
            ayuda="Se crea un documento nuevo que también debe firmarse; el consentimiento original no se modifica." />
          <div className="flex flex-wrap gap-3">
            <Boton type="submit" variante="peligro" disabled={accion.ocupado}>Preparar revocación</Boton>
            <Boton variante="secundario" onClick={() => setModo(null)}>Cancelar</Boton>
          </div>
        </form>
      )}
    </li>
  );
};

// Consentimientos informados del paciente (RF-06): formatos por procedimiento, firma en pantalla u OTP, PDF inmutable.
const PestanaConsentimientos = ({ pacienteId }: { pacienteId: number }) => {
  const { usuario } = useAuth();
  const esOdontologo = usuario?.rol === 'odontologo';
  const lista = useCarga(() => listarConsentimientos(pacienteId), `consentimientos-${pacienteId}`);
  const codigos = useCarga(() => listarCodigosEnviados(pacienteId), `codigos-${pacienteId}`);
  const catalogo = useCarga(() => listarProcedimientos(), 'procedimientos-catalogo');
  const accion = useAccion();
  const [procedimientoId, setProcedimientoId] = useState('');
  const [pieza, setPieza] = useState('');

  const preparar = async (e: FormEvent) => {
    e.preventDefault();
    if (await accion.ejecutar(() => crearConsentimiento(pacienteId, Number(procedimientoId), pieza ? Number(pieza) : undefined))) {
      setProcedimientoId('');
      setPieza('');
      lista.recargar();
    }
  };
  const recargar = () => { lista.recargar(); codigos.recargar(); };

  return (
    <div className="space-y-6">
      {esOdontologo && (
        <Tarjeta>
          <form onSubmit={preparar} className="space-y-4">
            <h2 className="text-seccion font-semibold">Preparar un consentimiento</h2>
            <p className="text-etiqueta text-texto-suave">
              Los datos del paciente, el procedimiento, el profesional y la fecha se completan solos. Si el paciente es menor de edad, firma su tutor legal.
              Un procedimiento marcado «requiere consentimiento» no se puede iniciar sin un consentimiento firmado y vigente.
            </p>
            <div className="grid gap-4 md:grid-cols-3">
              <Selector id="cons-procedimiento" etiqueta="Procedimiento" required value={procedimientoId} onChange={(e) => setProcedimientoId(e.target.value)}>
                <option value="">Elige…</option>
                {catalogo.datos?.filter((p) => p.activo).map((p) => <option key={p.id} value={p.id}>{p.nombre}{p.requiere_consentimiento ? ' · exige consentimiento' : ''}</option>)}
              </Selector>
              <Campo id="cons-pieza" etiqueta="Pieza FDI (opcional)" inputMode="numeric" maxLength={2} value={pieza} onChange={(e) => setPieza(e.target.value.replace(/\D/g, ''))} />
            </div>
            {accion.error && <Aviso>{accion.error}</Aviso>}
            <Boton type="submit" disabled={accion.ocupado || !procedimientoId}>Preparar consentimiento</Boton>
          </form>
        </Tarjeta>
      )}

      {lista.error && <Aviso>{lista.error}</Aviso>}
      {lista.datos?.length === 0 && <p className="text-texto-suave">El paciente aún no tiene consentimientos.</p>}
      <ul className="space-y-3">
        {lista.datos?.map((c) => <FilaConsentimiento key={`${c.id}-${c.estado}-${c.revocado}`} c={c} pacienteId={pacienteId} alCambiar={recargar} puedeRevocar={esOdontologo} />)}
      </ul>

      {codigos.datos && codigos.datos.length > 0 && (
        <Tarjeta denso className="space-y-3">
          <h2 className="text-seccion font-semibold">Códigos enviados</h2>
          <p className="text-etiqueta text-texto-suave">Con el adaptador de mensajes simulado, el código se ve aquí para poder probar; con un proveedor real aparece oculto.</p>
          <ul className="space-y-2 text-etiqueta">
            {codigos.datos.map((n) => (
              <li key={n.id}><span className="font-mono">{fechaHoraLocal(n.creado_en)}</span> · {n.canal} a {n.destino} · {n.estado}<br /><span className="text-texto-suave">{n.mensaje}</span></li>
            ))}
          </ul>
        </Tarjeta>
      )}
    </div>
  );
};

export default PestanaConsentimientos;
