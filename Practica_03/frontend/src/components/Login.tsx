// src/components/Login.tsx
import { useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth, type Rol } from "../context/AuthContext";
import { apiFetch } from "../services/http";
import InstalarApp from './InstalarApp';
import Logo from "./Logo";

const Login = () => {
  const [email, setEmail] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [error, setError] = useState<string>("");
  const [loading, setLoading] = useState<boolean>(false);

  const navigate = useNavigate();
  const { login } = useAuth();

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setLoading(true);

    apiFetch<{ token: string; email: string; rol: string }>("/api/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    })
      .then((data) => {
        setError("");
        const rol: Rol = data.rol === "admin" ? "admin" : "cliente";
        login({ email: data.email, rol }, data.token);

        // Redirigimos según el rol: admin al Dashboard, cliente a la Tienda
        navigate(rol === "admin" ? "/" : "/tienda");
      })
      .catch((err) => {
        setError(err.message || "Credenciales incorrectas");
      })
      .finally(() => {
        setLoading(false);
      });
  };

  const campo =
    "w-full border border-linea bg-papel px-4 py-3 text-tinta placeholder:text-piedra outline-none transition focus:border-marino-700 focus:ring-1 focus:ring-marino-700";
  const etiqueta = "mb-2 block text-xs font-semibold uppercase tracking-widest text-slate-600";

  return (
    <div className="min-h-screen flex items-center justify-center bg-fondo px-4">
      <div className="w-full max-w-md border border-linea border-t-[6px] border-t-marino-700 bg-papel p-8">
        <div className="mb-8 text-center">
          <Logo size={56} className="flex-col !gap-3" />
          <p className="mt-3 text-slate-600">Ingresa a tu cuenta para continuar</p>
        </div>

        {error && (
          <div className="mb-6 border border-coral-200 border-l-4 border-l-coral-500 bg-coral-50 p-3 text-center text-sm text-coral-700">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label className={etiqueta}>Correo electrónico</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className={campo}
              placeholder="admin@upse.edu.ec"
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
              required
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-marino-700 py-3 font-semibold uppercase tracking-wider text-arena-200 transition hover:bg-marino-800 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {loading ? "Validando..." : "Iniciar sesión"}
          </button>
        </form>

        <InstalarApp ubicacion="login" />

        <p className="mt-6 text-center text-sm text-slate-600">
          ¿Aún no tienes cuenta?{" "}
          <Link to="/registro" className="font-semibold text-coral-600 underline underline-offset-4 hover:text-coral-700">
            Crear cuenta
          </Link>
        </p>

        <div className="mt-6 space-y-1 border border-linea bg-arena-50 p-4 text-xs text-slate-600">
          <p className="font-semibold uppercase tracking-widest text-slate-700">Cuentas de prueba</p>
          <p>
            👑 Admin: <span className="font-mono">admin@upse.edu.ec / 123456</span>
          </p>
          <p>
            🛍️ Cliente: <span className="font-mono">cliente@upse.edu.ec / 123456</span>
          </p>
        </div>
      </div>
    </div>
  );
};

export default Login;
