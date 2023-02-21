import { useState, useEffect, useCallback } from 'react';
import { useMedia } from 'react-use';
// @ts-ignore - No Typescript types.
import * as Plot from '@observablehq/plot';
import {
    ApiResponse,
    OrderT,
    ProductT,
    RichOrderT,
    StationT,
    UserT,
} from '../../app/types';

export * from './useProductTypes';

interface IGetOrdersListState {
    loading: boolean;
    error: string | null;
    orders: RichOrderT[];
    page: number;
};

/**
 * Gets the list of orders.
 * @param page The page to get.
 * @returns The details.
 */
export const useGetOrdersList = (page: number) => {
    const [state, setState] = useState<IGetOrdersListState>({
        loading: false,
        error: null,
        orders: [],
        page,
    });

    /**
     * Fetches the list of orders.
     * @param orderId Optional parameter when searching for a specific order.
     * @returns The order(s) found.
     */
    const fetchOrders = async (orderId?: number) => {
        setState({
            ...state,
            loading: true,
            orders: [],
        });

        if (!orderId) {
            const req = await fetch(`/api/orders?p=${page}`);
            const resp = await req.json();

            if (resp.success === true) {
                setState({
                    ...state,
                    loading: false,
                    orders: resp.data || [],
                });

                return resp.data as RichOrderT[];
            } else {
                setState({
                    ...state,
                    loading: false,
                    error: resp.error || 'Unable to fetch orders',
                });

                return null;
            }
        } else {
            const req = await fetch(`/api/order/${orderId}`);
            const resp = await req.json();

            if (resp.success === true && !!resp.data.order.id) {
                setState({
                    ...state,
                    loading: false,
                    orders: [resp.data] || [],
                });

                return [resp.data] as RichOrderT[];
            } else {
                setState({
                    ...state,
                    loading: false,
                    error: resp.error,
                    orders: [],
                });

                return null;
            }
        }
    };

    useEffect(() => {
        fetchOrders();
    }, [page]);

    return { ...state, fetchOrders };
};

/**
 * A hook that gets the total earnings year-to-date.
 * @returns The total earnings.
 */
export const useGetTotalEarnings = () => {
    const [earnings, setEarnings] = useState(0);

    useEffect(() => {
        const getEarnings = async () => {
            const req = await fetch('/api/orders/earnings');
            const resp = await req.json();

            setEarnings(resp.data);
        };

        getEarnings();
    });

    return earnings;
};

/**
 * Gets the earnings for the past 4 days.
 * @returns The earnings for the past 4 days.
 */
export const useGetEarningsPerDay = () => {
    const [earnings, setEarnings] = useState<number[]>([0, 0, 0, 0, 0]);

    useEffect(() => {
        const getEarnings = async (day: number): Promise<number[]> => {
            const req = await fetch(`/api/orders/earnings/${day}`);
            const resp = await req.json();

            return [day, resp.data];
        };

        Promise.all([
            getEarnings(0),
            getEarnings(1),
            getEarnings(2),
            getEarnings(3),
            getEarnings(4),
        ]).then((queue) => {
            if (queue.length > 0) {
                const newEarnings = [...earnings];
    
                for (let [day, amount] of queue) {
                    newEarnings[day] = amount;
                }
    
                setEarnings(newEarnings);
            }
        });
    }, []);

    return earnings;
};

interface IGetProductListState {
    loading: boolean;
    error: string | null;
    products: ProductT[];
};

/**
 * Gets the list of all products.
 * @returns The details.
 */
 export const useGetProductsList = () => {
    const [state, setState] = useState<IGetProductListState>({
        loading: false,
        error: null,
        products: [],
    });

    /**
     * Fetches the list of products.
     */
    const fetchProducts = async () => {
        setState({
            ...state,
            loading: true,
            products: [],
        });

        const req = await fetch(`/api/products?all=1`);
        const resp = await req.json();

        if (resp.success === true) {
            setState({
                ...state,
                loading: false,
                products: resp.data,
            });

            return resp.data as ProductT[];
        } else {
            setState({
                ...state,
                loading: false,
                error: resp.error || 'Unable to fetch products',
            });

            return null;
        }
    };

    useEffect(() => {
        fetchProducts();
    }, []);

    return { ...state, fetchProducts };
};

/**
 * Returns whether we're in a compact (mobile) view.
 */
export const useIsCompactView = (): boolean => {
    return useMedia('(max-width: 500px)');
};

type UseStationsT =  {
    stations: StationT[];
    createStation: (name: string) => Promise<ApiResponse<StationT | null>>;
    getStations: () => Promise<ApiResponse<StationT[] | null>>;
    deleteStation: (id: number) => Promise<ApiResponse<null>>;
    addProductToStation: (sId: number, pId: number) => Promise<ApiResponse<null>>;
    removeProductFromStation: (sId: number, pId: number) => Promise<ApiResponse<null>>;
};

/**
 * This custom hook returns a list of stations that are in the DB. It also handles the
 * creation and deletion of stations.
 * @returns The stations and the function to manually get them.
 */
export const useStations = (shouldGetList = true): UseStationsT => {
    const [stations, setStations] = useState<StationT[]>([]);

    const getStations = useCallback(async () => {
        const req = await fetch('/api/stations');
        const resp = await req.json();

        if (resp.success) {
            setStations(resp.data);
        }

        return resp;
    }, []);

    const createStation = useCallback(async (name: string) => {
        const body = new FormData();
        body.append('name', name);

        const req = await fetch('/api/station', {
            method: 'POST',
            body,
        });
        const resp = await req.json();

        if (resp.success) {
            getStations();
        }

        return resp;
    }, []);

    const deleteStation = useCallback(async (id: number) => {
        const req = await fetch(`/api/station/${id}`, { method: 'DELETE' });
        const resp = await req.json();

        if (resp.success) {
            getStations();
        }

        return resp;
    }, []);

    const addProductToStation = useCallback(async (stationId: number, productId: number) => {;
        const req = await fetch(`/api/station/${stationId}/${productId}`, {
            method: 'POST',
        });
        const resp = await req.json();

        if (resp.success) {
            getStations();
        }

        return resp;
    }, []);

    const removeProductFromStation = useCallback(async (stationId: number, productId: number) => {;
        const req = await fetch(`/api/station/${stationId}/${productId}`, {
            method: 'DELETE',
        });
        const resp = await req.json();

        if (resp.success) {
            getStations();
        }

        return resp;
    }, []);

    useEffect(() => {
        if (!shouldGetList) {
            return;
        }

        getStations();
    }, []);

    return {
        stations,
        getStations,
        createStation,
        deleteStation,
        addProductToStation,
        removeProductFromStation,
    };
};

type UseUsersReturnT = {
    users: UserT[];
    createUser: (u: string, p: string, s?: number) => Promise<ApiResponse<null>>;
    getUsers: () => Promise<ApiResponse<UserT[]>>;
    deleteUser: (u: number) => Promise<ApiResponse<null>>;
};

/**
 * A hook that provides methods for managing users.
 */
export const useUsers = (): UseUsersReturnT => {
    const [users, setUsers] = useState<UserT[]>([]);

    const getUsers = useCallback(async () => {
        const req = await fetch('/api/users');
        const resp = await req.json();

        if (resp.success && !!resp.data) {
            setUsers(resp.data);
        }

        return resp;
    }, []);

    const createUser = useCallback(async (
        username: string,
        password: string,
        stationId = 0
    ) => {
        const body = new FormData();
        body.append('username', username);
        body.append('password', password);
        body.append('station_id', stationId.toString());

        const req = await fetch('/api/user', {
            method: 'POST',
            body,
        });
        const resp = await req.json();

        if (resp.success) {
            getUsers();
        }

        return resp;
    }, []);

    const deleteUser = useCallback(async (userId: number) => {;
        const req = await fetch(`/api/user/${userId}`, {
            method: 'DELETE',
        });
        const resp = await req.json();

        if (resp.success) {
            getUsers();
        }

        return resp;
    }, []);

    useEffect(() => {
        getUsers();
    }, []);

    return {
        users,
        createUser,
        getUsers,
        deleteUser,
    };
};

type OrdersStateT = {
    loading: boolean;
    error: string | null;
    orders: OrderT[];
};

/**
 * Gets the orders from the past year and also provides a function
 * to create an SVG graph of it.
 * @returns
 */
export const useGetOrdersPastYear = () => {
    const [state, setState] = useState<OrdersStateT>({
        loading: true,
        error: null,
        orders: [],
    });

    const getOrders = async () => {
        const req = await fetch(`/api/orders/past_year`);
        const resp = await req.json();

        if (resp.success === true) {
            const orders = (resp.data as OrderT[] || []).map((o) => {
                return {
                    ...o,
                    timeOfDay: (new Date(o.created_at)).getHours(),
                    productsLen: o.products.length,
                    total: o.products.map((p) => p.price).reduce((a, b) => a + b, 0),
                };
            });

            setState({
                ...state,
                loading: false,
                orders,
            });

            return resp.data as RichOrderT[];
        } else {
            setState({
                ...state,
                loading: false,
                error: resp.error || 'Unable to fetch orders',
            });

            return null;
        }
    };

    const getHexGraph = (): SVGElement => {
        return Plot.plot({
            grid: true,
            inset: 10,
            color: {
                scheme: 'RdYlBu'
            },
            marks: [
                Plot.hexagon(
                    state.orders,
                    Plot.hexbin({
                        fill: 'count',
                    }, {
                        x: 'timeOfDay',
                        y: 'total',
                        stroke: 'rgba(0, 0, 0, 0.1)',
                        strokeWidth: 1,
                    })
                )
            ]
        });
    };

    const getBoxOpaqueGraph = (): SVGElement => {
        return Plot.plot({
            y: {
                label: '↑ Amount of orders',
            },
            marks: [
                Plot.rectY(
                    state.orders,
                    Plot.binX({
                        y: 'count',
                    }, {
                        x: 'timeOfDay',
                        fill: '#1d54a8',
                    })
                ),
                Plot.ruleY([0])
            ]
        })
        // return Plot.rect(
        //     state.orders,
        //     Plot.bin({
        //         fillOpacity: 'count',
        //     }, {
        //         x: 'timeOfDay',
        //         y: 'total',
        //         fill: 'productsLen',
        //     })
        // ).plot();
    };

    useEffect(() => {
        getOrders();
    }, []);

    return {
        ...state,
        getOrders,
        getHexGraph,
        getBoxOpaqueGraph,
    };
};
