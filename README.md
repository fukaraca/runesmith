# Runesmith
Ancient dark magic meets modern orchestration
<p align="center">
  <img src="motivation.jpg" alt="Motivational Image" style="max-width:30%; height:auto;" />
</p>
A small Kubernetes-native demo that forges magical items using Mana.
It does not solve a real problem. It exists as showcase for Kubernetes operators/controllers, a custom Device Plugin, Kueue-style HPC workload scheduling, and some microservice basics end-to-end.

If you came here looking for business value, sorry. If you want a compact, working example of k8s-native patterns, this is it.

---
## Story in short
* Generate a random magical item by utilizing Kubernetes functionalities 
* Worker nodes are **anvil**. 
* Each anvil produces **essences** (Fire, Frost, Arcane) by spending **Mana**. Think mana as GPU resources
* An **Enchantment** CRD is the order form: “make this item with these essence amounts.” 
* The **Operator** watches Enchantments and orchestrates actual jobs. 
* **Kueue** keeps the line fair: scarce mana, priorities, preemption. 
* The **Device Plugin** lives on each anvil, giving it an identity and describing what kind of jobs it can take. 
* A tiny **backend** and **UI** sit on top so you can click a button and watch the system breathe.

---
## Components

* **Enchantment (CRD)** — the desired state. “Please craft X, needs Y Mana.”
* **Runesmith Operator** — the brain: watches Enchantments, spawns/updates jobs per energy type, writes status.
* **Manawell Device Plugin** — runs on each anvil as daemon-set; manages Mana as resource(eg like GPU compute unit)
* **Kueue** — the queue/scheduler layer that respects priority (Legendary > others) and available capacity.
* **Backend (Go)** — REST API to create,watch Enchantments over k8s api, provide status of nodes and artifacts
* **UI (React)** — click to generate, watch queues and completion.


* **CRD kind**: `Enchantment` (the desired item and its essence needs).
* **Anvils (nodes)**: specialized workers; each one is aligned to Fire/Frost/Arcane.
* **Mana**: resource of the anvil. 5 Frost essence created by 5 Mana by frost node.
* **Statuses**: `Queued → Scheduled → Running → (Preempted?) → Completed`.

---

## Example item generation flow end-to-end

From the tiny catalog:

```yaml
- id: 38
  name: "God-slayer Elemental Blade"
  tier: "Legendary"
  requirements:
    fire: 6
    frost: 5
    arcane: 4
  priority: 2
```

1. Legendary gets higher priority. Requirements say we need 6 essence of fire, 5 frost and 4 arcane
2. UI requests a random Item.
3. Backend generates one and orders Enchantment(crd) into etcd
4. Operator watches the orders and spawns Jobs (6x Fire, 5x Frost, 4x Arcane)
5. Kueue places the jobs accordingly, prioritize or preempts by its rules
6. Device plugin advertises its **resource type**(_manawell.io/fire_) and mana inventory to the kubelet.
7. Completion detected by backend and status shown in the UI

