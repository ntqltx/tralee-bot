const TELEGRAM_API = "https://api.telegram.org";

export default {
  async fetch(request, env) {
    if (request.method !== "POST") {
      return new Response("ok", { status: 200 });
    }

    let update;
    try {
      update = await request.json();
    } catch {
      return new Response("bad json", { status: 400 });
    }

    const msg = update?.message;
    const chatId = msg?.chat?.id;
    const text = (msg?.text || "").trim();

    if (!chatId || !text) {
      return new Response("ok", { status: 200 });
    }

    if (text === "/start") {
      await env.KV.put(`sub:${chatId}`, String(chatId));
      await reply(env, chatId, "✅ You've subscribed to Tralee apartment alerts! Send /stop to unsubscribe.");
    } else if (text === "/stop") {
      await env.KV.delete(`sub:${chatId}`);
      await reply(env, chatId, "❌ Unsubscribed. Send /start to resubscribe.");
    }

    return new Response("ok", { status: 200 });
  },
};

async function reply(env, chatId, text) {
  const url = `${TELEGRAM_API}/bot${env.BOT_TOKEN}/sendMessage`;
  await fetch(url, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify({ chat_id: chatId, text }),
  });
}
