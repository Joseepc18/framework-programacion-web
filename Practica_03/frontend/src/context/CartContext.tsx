// src/context/CartContext.tsx
import { createContext, useContext, useEffect, useState, type ReactNode } from 'react';
import type { Producto } from '../data/productos';
import { useAuth } from './AuthContext';

// 1. Elemento del carrito: un producto + la cantidad seleccionada
export interface CartItem extends Producto {
  cantidad: number;
}

interface CartContextType {
  cart: CartItem[];
  addToCart: (producto: Producto) => void;
  removeFromCart: (id: number) => void;
  incrementQuantity: (id: number) => void;
  decrementQuantity: (id: number) => void;
  clearCart: () => void;
  totalItems: number;
  totalPrice: number;
}

// 2. Prefijo de la clave de persistencia en localStorage.
// Se completa con el correo del usuario para que CADA CUENTA tenga su propio carrito.
const CART_STORAGE_PREFIX = "multicatalogo_carrito_";

// 3. Creamos el contexto indicando que puede ser CartContextType o undefined
const CartContext = createContext<CartContextType | undefined>(undefined);

// 4. Hook personalizado con validación de tipo
export const useCart = () => {
  const context = useContext(CartContext);
  if (!context) {
    throw new Error("useCart debe ser usado dentro de un CartProvider");
  }
  return context;
};

// 5. Tipamos los props del Provider
interface CartProviderProps {
  children: ReactNode;
}

// 6. El Provider
export const CartProvider = ({ children }: CartProviderProps) => {
  const { user } = useAuth();

  // Clave de almacenamiento según el usuario autenticado.
  // App.tsx remonta este provider con key={user?.email} al cambiar de cuenta,
  // por lo que el estado siempre se inicializa con el carrito de ESTE usuario.
  const storageKey = user
    ? `${CART_STORAGE_PREFIX}${user.email}`
    : `${CART_STORAGE_PREFIX}anonimo`;

  // Inicializamos el estado leyendo el carrito guardado del usuario actual
  const [cart, setCart] = useState<CartItem[]>(() => {
    try {
      const saved = localStorage.getItem(storageKey);
      return saved ? (JSON.parse(saved) as CartItem[]) : [];
    } catch {
      return [];
    }
  });

  // Persistimos el carrito del usuario en cada cambio (Tema 5: carrito persistente)
  useEffect(() => {
    localStorage.setItem(storageKey, JSON.stringify(cart));
  }, [cart, storageKey]);

  const addToCart = (producto: Producto) => {
    setCart((prevCart) => {
      const itemExists = prevCart.find((item) => item.id === producto.id);
      if (itemExists) {
        return prevCart.map((item) =>
          item.id === producto.id
            ? { ...item, cantidad: item.cantidad + 1 }
            : item
        );
      }
      return [...prevCart, { ...producto, cantidad: 1 }];
    });
  };

  const removeFromCart = (id: number) => {
    setCart((prevCart) => prevCart.filter((item) => item.id !== id));
  };

  const incrementQuantity = (id: number) => {
    setCart((prevCart) =>
      prevCart.map((item) =>
        item.id === id ? { ...item, cantidad: item.cantidad + 1 } : item
      )
    );
  };

  const decrementQuantity = (id: number) => {
    setCart((prevCart) =>
      prevCart
        .map((item) =>
          item.id === id ? { ...item, cantidad: item.cantidad - 1 } : item
        )
        .filter((item) => item.cantidad > 0) // si llega a 0, se elimina
    );
  };

  // Vacía el carrito tras confirmar el pedido
  const clearCart = () => {
    setCart([]);
  };

  const totalItems = cart.reduce((total, item) => total + item.cantidad, 0);
  const totalPrice = cart.reduce((total, item) => total + item.precio * item.cantidad, 0);

  return (
    <CartContext.Provider
      value={{
        cart,
        addToCart,
        removeFromCart,
        incrementQuantity,
        decrementQuantity,
        clearCart,
        totalItems,
        totalPrice,
      }}
    >
      {children}
    </CartContext.Provider>
  );
};