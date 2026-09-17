import { useState } from 'react';
import TarjetaUsuario from './components/TarjetaUsuario';
import PokemonView from './components/PokemonView';

interface Usuario {
  name: string;
  email: string;
  phone: string;
}

export default function App() {
  const [usuario, setUsuario] = useState<Usuario | null>(null);
  const [cargando, setCargando] = useState<boolean>(false);
  const [verPokedex, setVerPokedex] = useState<boolean>(false);

  const obtenerDatos = async () => {
    setCargando(true);
    try {
      const respuesta = await fetch('https://jsonplaceholder.typicode.com/users/1');
      const datos = await respuesta.json();
      setUsuario(datos);
    } catch (error) {
      console.error("Error al consumir la API", error);
    } finally {
      setCargando(false);
    }
  };

  if (verPokedex) {
    return <PokemonView onRegresar={() => setVerPokedex(false)} />;
  }

  return (
    <div className="p-8 bg-slate-100 min-h-screen">
      <h1 className="text-2xl font-bold mb-4 text-slate-800">Fundamentos de React</h1>
      
      <div className="flex flex-wrap items-center gap-3">
        <button
          onClick={obtenerDatos}
          className="bg-blue-600 text-white px-4 py-2 rounded shadow hover:bg-blue-700 transition-all cursor-pointer font-medium"
        >
          {cargando ? 'Consultando API...' : 'Obtener Usuario'}
        </button>

        <button
          onClick={() => setVerPokedex(true)}
          className="bg-green-600 text-white px-4 py-2 rounded shadow hover:bg-green-700 transition-all cursor-pointer font-medium"
        >
          Ver Pokédex
        </button>
      </div>

      {usuario && (
        <TarjetaUsuario nombre={usuario.name} correo={usuario.email} phone={usuario.phone} />
      )}
    </div>
  );
}




