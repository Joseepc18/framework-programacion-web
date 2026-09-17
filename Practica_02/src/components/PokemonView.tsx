import { useEffect, useState } from 'react';

interface Props {
  onRegresar: () => void;
}

interface PokemonListItem {
  name: string;
  url: string;
}

interface PokemonDetalle {
  id: number;
  name: string;
  height: number;
  weight: number;
  base_experience: number;
  sprites: {
    front_default: string | null;
    other: {
      'official-artwork': {
        front_default: string | null;
      };
    };
  };
  types: { type: { name: string } }[];
  abilities: { ability: { name: string } }[];
  stats: { base_stat: number; stat: { name: string } }[];
}

const NOMBRES_STATS: Record<string, string> = {
  hp: 'HP',
  attack: 'Ataque',
  defense: 'Defensa',
  'special-attack': 'Atq. Especial',
  'special-defense': 'Def. Especial',
  speed: 'Velocidad',
};

const TYPE_COLORS: Record<string, string> = {
  fire: 'bg-red-600 text-white',
  water: 'bg-blue-600 text-white',
  grass: 'bg-emerald-600 text-white',
  electric: 'bg-amber-500 text-black font-semibold',
  psychic: 'bg-pink-600 text-white',
  ice: 'bg-sky-400 text-black font-semibold',
  dragon: 'bg-indigo-600 text-white',
  ghost: 'bg-purple-800 text-white',
  poison: 'bg-purple-600 text-white',
  ground: 'bg-amber-700 text-white',
  rock: 'bg-stone-600 text-white',
  bug: 'bg-lime-600 text-white',
  fighting: 'bg-orange-600 text-white',
  fairy: 'bg-pink-400 text-white',
  steel: 'bg-slate-500 text-white',
};

const capitalizar = (texto: string) =>
  texto.charAt(0).toUpperCase() + texto.slice(1);

export default function PokemonView({ onRegresar }: Props) {
  const [lista, setLista] = useState<PokemonListItem[]>([]);
  const [busqueda, setBusqueda] = useState<string>('');
  const [seleccionado, setSeleccionado] = useState<string>('');
  const [detalle, setDetalle] = useState<PokemonDetalle | null>(null);

  useEffect(() => {
    const obtenerLista = async () => {
      try {
        const respuesta = await fetch('https://pokeapi.co/api/v2/pokemon?limit=151');
        const datos = await respuesta.json();
        setLista(datos.results);
      } catch (error) {
        console.error('Error al consumir la lista de pokémon', error);
      }
    };
    obtenerLista();
  }, []);

  useEffect(() => {
    if (!seleccionado) return;
    const obtenerDetalle = async () => {
      try {
        const respuesta = await fetch(`https://pokeapi.co/api/v2/pokemon/${seleccionado}`);
        const datos = await respuesta.json();
        setDetalle(datos);
      } catch (error) {
        console.error('Error al consumir el detalle', error);
      }
    };
    obtenerDetalle();
  }, [seleccionado]);

  const pokemonesFiltrados = lista.filter((p) =>
    p.name.toLowerCase().includes(busqueda.toLowerCase())
  );

  if (lista.length === 0) {
    return (
      <div className="min-h-screen bg-slate-950 flex flex-col items-center justify-center p-6 text-white font-sans">
        <div className="bg-slate-900 border border-red-600 p-8 rounded-xl text-center max-w-sm w-full shadow-lg">
          <div className="w-8 h-8 rounded-full border-2 border-red-500 border-t-transparent animate-spin mx-auto mb-4"></div>
          <p className="text-sm font-semibold tracking-wide text-slate-200">Cargando Pokédex...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 p-4 sm:p-6 font-sans selection:bg-red-500 selection:text-white">
      {/* Pokédex Outer Housing */}
      <div className="max-w-4xl mx-auto bg-slate-900 border border-slate-800 rounded-2xl p-4 sm:p-6 shadow-2xl">
        
        {/* Console Header Bar */}
        <div className="flex items-center justify-between border-b border-slate-800 pb-4 mb-6">
          <h1 className="text-base font-bold tracking-widest text-slate-200 font-mono uppercase">
            Pokédex
          </h1>

          <button
            onClick={onRegresar}
            className="bg-slate-800 hover:bg-slate-700 border border-slate-700 px-3.5 py-1.5 rounded-lg text-xs font-medium text-slate-200 transition-all cursor-pointer"
          >
            ← Regresar
          </button>
        </div>

        {/* Display Screen */}
        <div className="bg-slate-950 border border-slate-800 rounded-xl p-5 mb-6">
          {!detalle ? (
            <div className="min-h-[220px] flex flex-col items-center justify-center text-center p-6 text-slate-500">
              <p className="text-sm text-slate-400">Selecciona un Pokémon de la lista</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-12 gap-6 items-center">
              
              {/* Pokémon Artwork & Title */}
              <div className="md:col-span-5 flex flex-col items-center justify-center bg-slate-900/60 border border-slate-800 p-5 rounded-xl relative group">
                <span className="absolute top-3 left-3 font-mono text-xs text-slate-400">
                  #{String(detalle.id).padStart(3, '0')}
                </span>
                
                <span className="absolute top-3 right-3 font-mono text-xs text-amber-400 font-semibold">
                  {detalle.base_experience} XP
                </span>

                <img
                  src={detalle.sprites.other['official-artwork'].front_default ?? ''}
                  alt={detalle.name}
                  className="w-36 h-36 object-contain my-3 animate-float transition-transform duration-300 hover:scale-110 cursor-pointer"
                />

                <h2 className="text-2xl font-bold text-white tracking-wide text-center">
                  {capitalizar(detalle.name)}
                </h2>

                <div className="flex flex-wrap justify-center gap-1.5 mt-3">
                  {detalle.types.map((t) => {
                    const color = TYPE_COLORS[t.type.name.toLowerCase()] || 'bg-slate-800 text-white';
                    return (
                      <span key={t.type.name} className={`px-2.5 py-0.5 rounded text-xs capitalize ${color}`}>
                        {t.type.name}
                      </span>
                    );
                  })}
                </div>
              </div>

              {/* Specs & Stats */}
              <div className="md:col-span-7 flex flex-col space-y-4">
                
                {/* Physical Info */}
                <div className="grid grid-cols-3 gap-3 text-center">
                  <div className="bg-slate-900 border border-slate-800 p-2.5 rounded-lg">
                    <span className="text-[10px] text-slate-400 block uppercase font-mono">Altura</span>
                    <span className="text-sm font-semibold text-slate-200">{detalle.height / 10} m</span>
                  </div>
                  <div className="bg-slate-900 border border-slate-800 p-2.5 rounded-lg">
                    <span className="text-[10px] text-slate-400 block uppercase font-mono">Peso</span>
                    <span className="text-sm font-semibold text-slate-200">{detalle.weight / 10} kg</span>
                  </div>
                  <div className="bg-slate-900 border border-slate-800 p-2.5 rounded-lg">
                    <span className="text-[10px] text-slate-400 block uppercase font-mono">Exp. Base</span>
                    <span className="text-sm font-semibold text-amber-400">{detalle.base_experience}</span>
                  </div>
                </div>

                {/* Abilities */}
                <div className="bg-slate-900 border border-slate-800 p-3 rounded-lg">
                  <span className="text-xs font-semibold text-slate-400 block mb-2 font-mono uppercase">Habilidades</span>
                  <div className="flex flex-wrap gap-1.5">
                    {detalle.abilities.map((a) => (
                      <span key={a.ability.name} className="bg-slate-950 border border-slate-800 text-slate-300 px-2.5 py-1 rounded text-xs capitalize">
                        {a.ability.name.replace('-', ' ')}
                      </span>
                    ))}
                  </div>
                </div>

                {/* Base Stats */}
                <div className="bg-slate-900 border border-slate-800 p-3 rounded-lg">
                  <span className="text-xs font-semibold text-slate-400 block mb-2 font-mono uppercase">Estadísticas</span>
                  <div className="space-y-2">
                    {detalle.stats.map((s) => {
                      const percentage = Math.min(100, (s.base_stat / 150) * 100);
                      return (
                        <div key={s.stat.name}>
                          <div className="flex justify-between text-xs font-mono text-slate-300 mb-0.5">
                            <span className="text-slate-400">{NOMBRES_STATS[s.stat.name] ?? s.stat.name}</span>
                            <span className="font-semibold text-slate-200">{s.base_stat}</span>
                          </div>
                          <div className="h-1.5 w-full rounded-full bg-slate-950 overflow-hidden">
                            <div
                              className="h-full bg-blue-500 rounded-full"
                              style={{ width: `${percentage}%` }}
                            />
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>

              </div>

            </div>
          )}
        </div>

        {/* Bottom Catalog & Search */}
        <div className="bg-slate-900 border border-slate-800 rounded-xl p-4">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-3 mb-3">
            <span className="text-xs font-semibold text-slate-400 font-mono uppercase">
              Lista de Pokémon ({pokemonesFiltrados.length})
            </span>

            <input
              type="text"
              value={busqueda}
              onChange={(e) => setBusqueda(e.target.value)}
              placeholder="Buscar..."
              className="w-full sm:w-56 bg-slate-950 border border-slate-800 rounded-lg px-3 py-1.5 text-xs text-slate-200 placeholder:text-slate-600 outline-none focus:border-slate-600"
            />
          </div>

          <div className="max-h-48 overflow-y-auto pr-1">
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-2">
              {pokemonesFiltrados.map((pokemon) => {
                const activo = pokemon.name === seleccionado;
                return (
                  <button
                    key={pokemon.name}
                    onClick={() => setSeleccionado(pokemon.name)}
                    className={`px-3 py-2 rounded-lg text-xs font-medium transition-all cursor-pointer text-center truncate border ${
                      activo
                        ? 'bg-blue-600 text-white font-semibold border-blue-500'
                        : 'bg-slate-950 text-slate-300 border-slate-800 hover:bg-slate-800 hover:text-white'
                    }`}
                  >
                    {capitalizar(pokemon.name)}
                  </button>
                );
              })}
            </div>
          </div>
        </div>

      </div>
    </div>
  );
}