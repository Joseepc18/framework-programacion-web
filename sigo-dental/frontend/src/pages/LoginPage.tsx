import { useState, type FormEvent } from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth, type AvisoSesion } from '../context/useAuth';
import { ApiError } from '../services/http';

const AVISOS: Record<AvisoSesion, string> = {
  inactividad: 'Se cerró tu sesión por inactividad. Inicia sesión de nuevo para continuar.',
  expirada: 'Tu sesión expiró. Inicia sesión de nuevo para continuar.',
};

// Un mismo mensaje para cualquier 401: no se revela si falló el correo o la contraseña.
const mensajeDeError = (error: unknown) => {
  if (error instanceof ApiError) {
    return error.status === 401 || error.status === 400
      ? 'Correo o contraseña incorrectos. Revisa tus datos e inténtalo de nuevo.'
      : 'No se pudo iniciar sesión por un problema del servidor. Inténtalo de nuevo en unos minutos.';
  }
  return 'No se pudo conectar con la API. Verifica tu conexión e inténtalo de nuevo.';
};

const CAMPO =
  'mt-2 min-h-11 w-full rounded-control border border-chip-neutro-borde bg-chip-neutro-fondo px-4 text-cuerpo text-texto focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-acento';

const LoginPage = () => {
  const { usuario, aviso, iniciarSesion, descartarAviso } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [enviando, setEnviando] = useState(false);
  const [error, setError] = useState('');

  if (usuario) return <Navigate to="/" replace />;

  const enviar = async (evento: FormEvent<HTMLFormElement>) => {
    evento.preventDefault();
    setEnviando(true);
    setError('');
    descartarAviso();
    try {
      await iniciarSesion(email, password);
    } catch (fallo) {
      setError(mensajeDeError(fallo));
      setPassword('');
      setEnviando(false);
    }
  };

  return (
    <main className="grid min-h-dvh place-items-center p-4">
      <div className="vidrio w-full max-w-md rounded-tarjeta p-6 md:p-8">
        <p className="text-seccion font-semibold">
          SIGO<span className="text-acento">-Dental</span>
        </p>
        <h1 className="mt-4 text-pagina font-semibold">Iniciar sesión</h1>
        <p className="mt-2 text-texto-suave">Ingresa con tu cuenta de la clínica.</p>

        <div aria-live="polite">
          {aviso && !error && (
            <p className="mt-5 rounded-control border border-chip-neutro-borde bg-chip-neutro-fondo p-3 text-etiqueta text-chip-neutro-texto">
              {AVISOS[aviso]}
            </p>
          )}
          {error && (
            <p role="alert" className="mt-5 rounded-control border border-chip-alerta-borde bg-chip-alerta-fondo p-3 text-etiqueta text-chip-alerta-texto">
              {error}
            </p>
          )}
        </div>

        <form onSubmit={enviar} className="mt-6 space-y-5">
          <div>
            <label htmlFor="correo" className="text-etiqueta font-medium text-texto-suave">Correo electrónico</label>
            <input
              id="correo"
              type="email"
              autoComplete="username"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className={CAMPO}
            />
          </div>
          <div>
            <label htmlFor="contrasena" className="text-etiqueta font-medium text-texto-suave">Contraseña</label>
            <input
              id="contrasena"
              type="password"
              autoComplete="current-password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className={CAMPO}
            />
          </div>
          <button
            type="submit"
            disabled={enviando}
            className="inline-flex min-h-11 w-full items-center justify-center rounded-control bg-acento px-5 text-cuerpo font-semibold text-sobre-acento transition hover:brightness-110 disabled:opacity-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-acento"
          >
            {enviando ? 'Verificando…' : 'Iniciar sesión'}
          </button>
        </form>
      </div>
    </main>
  );
};

export default LoginPage;
