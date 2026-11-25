import { redirect } from "next/navigation"

let accessToken: string | null = null
let refreshPromise: Promise<string> | null = null


async function refreshAccessToken(): Promise<string> {

    if (refreshPromise) {
        return refreshPromise
    }

    refreshPromise = (async () => {
        const response = await fetch("/auth/refresh", {
            "method": "POST",
            credentials: "include",
        })

        if (!response.ok) {
            console.log(response)
            window.location.replace("/login")
        }

        accessToken = await response.text()

        return accessToken
    })()

    try {
        return await refreshPromise
    } finally {
        refreshPromise = null
    }

}


export async function callHome(
    url: string,
    options: RequestInit = {},
): Promise<Response> {

    const token = accessToken

    const send = (token: string | null) => {

        const headers = new Headers(options.headers);

        if (token) {
            headers.set("Authorization", `Bearer ${token}`)
        } else {
            headers.delete("Authorization")
        }

        const response = fetch(url, {
            ...options,
            headers,
        })

        return response
    }

    let response = await send(token)

    if (response.status !== 401) {
        return response
    }

    const newToken =
        token !== accessToken && accessToken
            ? accessToken
            : await refreshAccessToken()

    response = await send(newToken);

    return response
}
