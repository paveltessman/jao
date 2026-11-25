"use client"
import { callHome } from "@/lib/call-home"
import { useEffect, useState } from "react"

export function ContainerList() {
    const [state, setState] = useState("loading");

    useEffect(() => {
        callHome("/api/v1/containers")
            .then((response) => response.text())
            .then((response) => setState(response))
            .catch((err) => console.log(err))
    })

    return <p>{state}</p>
}
