import { useState, type FormEvent } from 'react';
import { Link, Navigate } from 'react-router-dom';
import { Aviso, Boton, Campo } from '../components/ui';
import { useAuth } from '../context/useAuth';
import { useAccion } from '../hooks/useAccion';
import { ingresarAlPortal, solicitarCodigo } from '../services/portalService';

// Ingreso del paciente al portal: documento de identidad y código de un solo uso enviado a su teléfono o correo.
const PortalIngresoPage = () => {
  const { usuario, establecerSesion } = useAuth();
  const accion = useAccion();
  const [documento, setDocumento] = useState('');
  const [codigo, setCodigo] = useState('');
  const [enviado, setEnviado] = useState('');

  if (usuario) return <Navigate to="/" replace />;

  const pedirCodigo = async (e: FormEvent) => {
    e.preventDefault();
    let mensaje = '';
    if (await accion.ejecutar(async () => { mensaje = (await solicitarCodigo(documento)).mensaje; })) setEnviado(mensaje);
  };

  const entrar = async (e: FormEvent) => {
    e.preventDefault();
    await accion.ejecutar(async () => establecerSesion(await ingresarAlPortal(documento, codigo)));
  };

  return (
    <main className="grid min-h-dvh place-items-center p-4">
      <div className="vidrio w-full max-w-md rounded-tarjeta p-6 md:p-8">
        <p className="text-seccion font-semibold">SIGO<span className="text-acento">-Dental</span></p>
        <h1 className="mt-4 text-pagina font-semibold">Portal del paciente</h1>
        <p className="mt-2 text-texto-suave">Consulta tus citas, recetas y consentimientos desde tu teléfono.</p>

        <form onSubmit={enviado ? entrar : pedirCodigo} className="mt-6 space-y-5">
          <Campo id="portal-documento" etiqueta="Documento de identidad" inputMode="numeric" autoComplete="off" required maxLength={30}
            value={documento} onChange={(e) => setDocumento(e.target.value)} disabled={!!enviado} />
          {enviado && (
            <>
              <Aviso tipo="info">{enviado}</Aviso>
              <Campo id="portal-codigo" etiqueta="Código de 6 dígitos" inputMode="numeric" autoComplete="one-time-code" required maxLength={6}
                value={codigo} onChange={(e) => setCodigo(e.target.value.replace(/\D/g, ''))} />
            </>
          )}
          {accion.error && <Aviso>{accion.error}</Aviso>}
          <div className="flex flex-wrap gap-3">
            <Boton type="submit" className="w-full" disabled={accion.ocupado || (enviado !== '' && codigo.length !== 6)}>
              {enviado ? 'Ingresar' : 'Enviarme un código'}
            </Boton>
            {enviado && (
              <Boton variante="secundario" className="w-full" onClick={() => { setEnviado(''); setCodigo(''); accion.limpiar(); }}>Cambiar documento o pedir otro código</Boton>
            )}
          </div>
        </form>
        <p className="mt-6 text-etiqueta text-texto-suave">¿Eres parte del personal de la clínica? <Link className="text-acento underline" to="/login">Inicia sesión aquí</Link>.</p>
      </div>
    </main>
  );
};

export default PortalIngresoPage;
