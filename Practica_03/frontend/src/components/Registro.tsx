// src/components/Registro.tsx
// Alta pública de cuentas de cliente (Fase 1 del back-office).
// Si la URL trae ?ref=CODIGO, el campo viene prellenado y la cuenta nueva
// queda colgada del patrocinador en la red multinivel.
import { useState, type FormEvent } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { useAuth, type Rol } from "../context/AuthContext";
import { apiFetch } from "../services/http";
import Logo from "./Logo";

const Registro = () => {
  const [nombre, setNombre] = useState<string>("");
  const [email, setEmail] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [error, setError] = useState<string>("");
  const [loading, setLoading] = useState<boolean>(false);

  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { login } = useAuth();

  // Prefill del código de invitación desde la URL (/registro?ref=LUIS).
  const refInicial = (searchParams.get("ref") ?? "").toUpperCase();
  const [ref, setRef] = useState<string>(refInicial);

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setLoading(true);

    apiFetch<{ token?: string; message: string }>("/api/register", {
      method: "POST",
      body: JSON.stringify({
        nombre: nombre.trim(),
        email: email.trim(),
        password,
        ref: ref.trim().toUpperCase(),
      }),
    })
      .then(() => {
        // Tras registrar, iniciamos sesión automáticamente con las mismas credenciales.
        return apiFetch<{ token: string; email: string; rol: string }>("/api/login", {
          method: "POST",
          body: JSON.stringify({ email: email.trim(), password }),
        });
      })
      .then((sesion) => {
        setError("");
        const rol: Rol = sesion.rol === "admin" ? "admin" : "cliente";
        login({ email: sesion.email, rol }, sesion.token);
        navigate(rol === "admin" ? "/" : "/tienda");
      })
      .catch((err) => {
        setError(err.message || "No se pudo crear la cuenta");
      })
      .finally(() => {
        setLoading(false);
      });
  };

  const campo =
    "w-full border border-linea bg-white px-4 py-3 text-tinta placeholder:text-piedra outline-none transition focus:border-marino-600 focus:ring-1 focus:ring-marino-600";
  const etiqueta = "mb-2 block text-xs font-semibold uppercase tracking-widest text-slate-600";

  return (
    <div className="grid min-h-screen bg-fondo md:grid-cols-2">
      <aside className="hidden flex-col justify-between bg-marino-700 p-12 text-arena-200 md:flex">
        <Logo tono="claro" size={44} />
        <div>
          <h1 className="max-w-md font-display text-5xl leading-tight">
            Una cuenta, tu propia red.
          </h1>
          <p className="mt-6 max-w-md text-marino-200">
            Si llegas con un código de invitación, tu cuenta queda en la red de quien
            te invitó y ambos ganan con cada compra.
          </p>
        </div>
        <p className="text-xs uppercase tracking-widest text-marino-300">Jose Ponce Carlos</p>
      </aside>

      <main className="flex items-center justify-center px-6 py-12">
        <div className="w-full max-w-sm">
          <div className="mb-10 md:hidden">
            <Logo />
          </div>

          <h2 className="font-display text-4xl text-tinta">Crear cuenta</h2>
          <p className="mt-2 text-slate-600">Empieza a comprar en JP Mercantil.</p>

          {error && (
            <div className="mt-6 border-l-4 border-coral-500 bg-coral-50 p-3 text-sm text-coral-700">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="mt-8 space-y-5">
            <div>
              <label className={etiqueta}>Nombre completo</label>
              <input
                type="text"
                value={nombre}
                onChange={(e) => setNombre(e.target.value)}
                className={campo}
                placeholder="Ana García"
                required
              />
            </div>

            <div>
              <label className={etiqueta}>Correo electrónico</label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className={campo}
                placeholder="nombre@upse.edu.ec"
                required
              />
            </div>

            <div>
              <label className={etiqueta}>Contraseña</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className={campo}
                placeholder="••••••"
                minLength={6}
                required
              />
            </div>

            <div>
              <label className={etiqueta}>
                Código de referido <span className="font-normal normal-case tracking-normal text-slate-400">(opcional)</span>
              </label>
              <input
                type="text"
                value={ref}
                onChange={(e) => setRef(e.target.value.toUpperCase())}
                className={`${campo} font-mono uppercase`}
                placeholder="LUIS"
              />
              {refInicial && (
                <p className="mt-1 text-xs text-marino-600">
                  Invitado por el código {refInicial}
                </p>
              )}
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-marino-700 py-3 font-semibold text-arena-200 transition hover:bg-marino-800 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {loading ? "Creando cuenta..." : "Crear cuenta"}
            </button>
          </form>

          <p className="mt-6 text-sm text-slate-600">
            ¿Ya tienes cuenta?{" "}
            <Link to="/login" className="font-semibold text-coral-600 underline underline-offset-4 hover:text-coral-700">
              Inicia sesión
            </Link>
          </p>
        </div>
      </main>
    </div>
  );
};

export default Registro;
